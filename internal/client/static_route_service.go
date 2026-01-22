package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// StaticRouteService handles static route operations
type StaticRouteService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewStaticRouteService creates a new static route service instance
func NewStaticRouteService(executor Executor, client *rtxClient) *StaticRouteService {
	return &StaticRouteService{
		executor: executor,
		client:   client,
	}
}

// CreateRoute creates a new static route
func (s *StaticRouteService) CreateRoute(ctx context.Context, route StaticRoute) error {
	// Convert to parser type and validate
	parserRoute := s.toParserRoute(route)
	if err := parsers.ValidateStaticRoute(parserRoute); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Each next hop becomes a separate command
	for _, hop := range route.NextHops {
		parserHop := s.toParserHop(hop)
		cmd := parsers.BuildIPRouteCommand(parserRoute, parserHop)
		logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Creating static route with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create static route: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("route created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetRoute retrieves a static route configuration
func (s *StaticRouteService) GetRoute(ctx context.Context, prefix, mask string) (*StaticRoute, error) {
	cmd := parsers.BuildShowSingleRouteConfigCommand(prefix, mask)
	logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Getting static route with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get static route: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Static route raw output: %q", string(output))

	parser := parsers.NewStaticRouteParser()
	parserRoute, err := parser.ParseSingleRoute(string(output), prefix, mask)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static route: %w", err)
	}

	route := s.fromParserRoute(*parserRoute)
	return &route, nil
}

// UpdateRoute updates an existing static route
// This is done by deleting all existing next hops and recreating them
func (s *StaticRouteService) UpdateRoute(ctx context.Context, route StaticRoute) error {
	// Convert to parser type and validate
	parserRoute := s.toParserRoute(route)
	if err := parsers.ValidateStaticRoute(parserRoute); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current route configuration
	currentRoute, err := s.GetRoute(ctx, route.Prefix, route.Mask)
	if err != nil {
		// If route doesn't exist, just create it
		if strings.Contains(err.Error(), "not found") {
			return s.CreateRoute(ctx, route)
		}
		return fmt.Errorf("failed to get current route: %w", err)
	}

	// Delete all existing next hops
	for _, hop := range currentRoute.NextHops {
		parserHop := s.toParserHop(hop)
		cmd := parsers.BuildDeleteIPRouteCommand(route.Prefix, route.Mask, &parserHop)
		logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Deleting old next hop with command: %s", cmd)

		_, err := s.executor.Run(ctx, cmd)
		if err != nil {
			logging.FromContext(ctx).Warn().Str("service", "static_route").Msgf("Failed to delete old next hop: %v", err)
		}
	}

	// Create all new next hops
	for _, hop := range route.NextHops {
		parserHop := s.toParserHop(hop)
		cmd := parsers.BuildIPRouteCommand(parserRoute, parserHop)
		logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Creating new next hop with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create next hop: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("route updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteRoute removes a static route
func (s *StaticRouteService) DeleteRoute(ctx context.Context, prefix, mask string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current route to find all next hops
	currentRoute, err := s.GetRoute(ctx, prefix, mask)
	if err != nil {
		// If route doesn't exist, nothing to delete
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to get current route: %w", err)
	}

	// Delete all next hops
	for _, hop := range currentRoute.NextHops {
		parserHop := s.toParserHop(hop)
		cmd := parsers.BuildDeleteIPRouteCommand(prefix, mask, &parserHop)
		logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Deleting static route with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete static route: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			// Check if it's already gone
			if strings.Contains(strings.ToLower(string(output)), "not found") {
				continue
			}
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("route deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListRoutes retrieves all static routes
func (s *StaticRouteService) ListRoutes(ctx context.Context) ([]StaticRoute, error) {
	cmd := parsers.BuildShowIPRouteConfigCommand()
	logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Listing static routes with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list static routes: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "static_route").Msgf("Static routes raw output: %q", string(output))

	parser := parsers.NewStaticRouteParser()
	parserRoutes, err := parser.ParseRouteConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse static routes: %w", err)
	}

	// Convert parser routes to client routes
	routes := make([]StaticRoute, len(parserRoutes))
	for i, pr := range parserRoutes {
		routes[i] = s.fromParserRoute(pr)
	}

	return routes, nil
}

// toParserRoute converts client.StaticRoute to parsers.StaticRoute
func (s *StaticRouteService) toParserRoute(route StaticRoute) parsers.StaticRoute {
	nextHops := make([]parsers.NextHop, len(route.NextHops))
	for i, h := range route.NextHops {
		nextHops[i] = parsers.NextHop{
			NextHop:   h.NextHop,
			Interface: h.Interface,
			Distance:  h.Distance,
			Name:      h.Name,
			Permanent: h.Permanent,
			Filter:    h.Filter,
		}
	}

	return parsers.StaticRoute{
		Prefix:   route.Prefix,
		Mask:     route.Mask,
		NextHops: nextHops,
	}
}

// toParserHop converts client.StaticRouteHop to parsers.NextHop
func (s *StaticRouteService) toParserHop(hop StaticRouteHop) parsers.NextHop {
	return parsers.NextHop{
		NextHop:   hop.NextHop,
		Interface: hop.Interface,
		Distance:  hop.Distance,
		Name:      hop.Name,
		Permanent: hop.Permanent,
		Filter:    hop.Filter,
	}
}

// fromParserRoute converts parsers.StaticRoute to client.StaticRoute
func (s *StaticRouteService) fromParserRoute(pr parsers.StaticRoute) StaticRoute {
	nextHops := make([]StaticRouteHop, len(pr.NextHops))
	for i, h := range pr.NextHops {
		nextHops[i] = StaticRouteHop{
			NextHop:   h.NextHop,
			Interface: h.Interface,
			Distance:  h.Distance,
			Name:      h.Name,
			Permanent: h.Permanent,
			Filter:    h.Filter,
		}
	}

	return StaticRoute{
		Prefix:   pr.Prefix,
		Mask:     pr.Mask,
		NextHops: nextHops,
	}
}
