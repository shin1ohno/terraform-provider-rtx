package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// staticIntPtr is a helper function to create *int pointers for test data
func staticIntPtr(i int) *int {
	return &i
}

func TestNATStaticService_Create(t *testing.T) {
	tests := []struct {
		name        string
		nat         NATStatic
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful 1:1 NAT creation",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasType := false
					hasEntry := false
					for _, cmd := range cmds {
						if cmd == "nat descriptor type 1 static" {
							hasType = true
						}
						if cmd == "nat descriptor static 1 203.0.113.100=192.168.1.100" {
							hasEntry = true
						}
					}
					return hasType && hasEntry
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful port-based NAT creation",
			nat: NATStatic{
				DescriptorID: 2,
				Entries: []NATStaticEntry{
					{
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   staticIntPtr(8080),
						OutsideGlobal:     "203.0.113.100",
						OutsideGlobalPort: staticIntPtr(80),
						Protocol:          "tcp",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasType := false
					hasEntry := false
					for _, cmd := range cmds {
						if cmd == "nat descriptor type 2 static" {
							hasType = true
						}
						if cmd == "nat descriptor static 2 203.0.113.100:80=192.168.1.100:8080 tcp" {
							hasEntry = true
						}
					}
					return hasType && hasEntry
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful multiple entries creation",
			nat: NATStatic{
				DescriptorID: 3,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
					{
						InsideLocal:   "192.168.1.101",
						OutsideGlobal: "203.0.113.101",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasType := false
					hasEntry1 := false
					hasEntry2 := false
					for _, cmd := range cmds {
						if cmd == "nat descriptor type 3 static" {
							hasType = true
						}
						if cmd == "nat descriptor static 3 203.0.113.100=192.168.1.100" {
							hasEntry1 = true
						}
						if cmd == "nat descriptor static 3 203.0.113.101=192.168.1.101" {
							hasEntry2 = true
						}
					}
					return hasType && hasEntry1 && hasEntry2
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - invalid descriptor ID",
			nat: NATStatic{
				DescriptorID: 0,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid NAT static",
		},
		{
			name: "Validation error - missing inside local",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid NAT static",
		},
		{
			name: "Validation error - invalid IP address",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "invalid-ip",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid NAT static",
		},
		{
			name: "Execution error on type command",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "failed to create NAT static",
		},
		{
			name: "Command error with error output",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return([]byte("Error: descriptor already exists"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATStaticService{executor: mockExecutor}
			err := service.Create(context.Background(), tt.nat)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATStaticService_Get(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		mockSetup    func(*MockExecutor)
		expected     *NATStatic
		expectedErr  bool
		errMessage   string
	}{
		{
			name:         "Successful get 1:1 NAT",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 1 static
nat descriptor static 1 203.0.113.100=192.168.1.100
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*1"`).
					Return([]byte(output), nil)
			},
			expected: &NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			expectedErr: false,
		},
		{
			name:         "Successful get port-based NAT",
			descriptorID: 2,
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 2 static
nat descriptor static 2 203.0.113.100:80=192.168.1.100:8080 tcp
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*2"`).
					Return([]byte(output), nil)
			},
			expected: &NATStatic{
				DescriptorID: 2,
				Entries: []NATStaticEntry{
					{
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   staticIntPtr(8080),
						OutsideGlobal:     "203.0.113.100",
						OutsideGlobalPort: staticIntPtr(80),
						Protocol:          "tcp",
					},
				},
			},
			expectedErr: false,
		},
		{
			name:         "Successful get multiple entries",
			descriptorID: 3,
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 3 static
nat descriptor static 3 203.0.113.100=192.168.1.100
nat descriptor static 3 203.0.113.101=192.168.1.101
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*3"`).
					Return([]byte(output), nil)
			},
			expected: &NATStatic{
				DescriptorID: 3,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
					{
						InsideLocal:   "192.168.1.101",
						OutsideGlobal: "203.0.113.101",
					},
				},
			},
			expectedErr: false,
		},
		{
			name:         "NAT not found",
			descriptorID: 99,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*99"`).
					Return([]byte(""), nil)
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "NAT descriptor 99 not found",
		},
		{
			name:         "Execution error",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*1"`).
					Return(nil, errors.New("connection failed"))
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "failed to get NAT static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATStaticService{executor: mockExecutor}
			result, err := service.Get(context.Background(), tt.descriptorID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATStaticService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		mockSetup    func(*MockExecutor)
		expectedErr  bool
		errMessage   string
	}{
		{
			name:         "Successful deletion",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:         "Execution error",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "failed to delete NAT static",
		},
		{
			name:         "Already deleted (not found)",
			descriptorID: 99,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 99"}).
					Return([]byte("Error: not found"), nil)
			},
			expectedErr: false, // Should not error if already gone
		},
		{
			name:         "Command error with error output",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return([]byte("Error: descriptor in use"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATStaticService{executor: mockExecutor}
			err := service.Delete(context.Background(), tt.descriptorID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATStaticService_List(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    []NATStatic
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful list with multiple descriptors",
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 1 static
nat descriptor static 1 203.0.113.100=192.168.1.100
nat descriptor type 2 static
nat descriptor static 2 203.0.113.200:80=192.168.2.100:8080 tcp
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
					Return([]byte(output), nil)
			},
			expected: []NATStatic{
				{
					DescriptorID: 1,
					Entries: []NATStaticEntry{
						{
							InsideLocal:   "192.168.1.100",
							OutsideGlobal: "203.0.113.100",
						},
					},
				},
				{
					DescriptorID: 2,
					Entries: []NATStaticEntry{
						{
							InsideLocal:       "192.168.2.100",
							InsideLocalPort:   staticIntPtr(8080),
							OutsideGlobal:     "203.0.113.200",
							OutsideGlobalPort: staticIntPtr(80),
							Protocol:          "tcp",
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			name: "Empty list",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
					Return([]byte(""), nil)
			},
			expected:    []NATStatic{},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
					Return(nil, errors.New("connection failed"))
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "failed to list NAT statics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATStaticService{executor: mockExecutor}
			result, err := service.List(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				// Since map iteration is not deterministic, compare length and contents
				assert.Len(t, result, len(tt.expected))
				for _, expectedNAT := range tt.expected {
					found := false
					for _, resultNAT := range result {
						if resultNAT.DescriptorID == expectedNAT.DescriptorID {
							found = true
							assert.Equal(t, len(expectedNAT.Entries), len(resultNAT.Entries))
							break
						}
					}
					assert.True(t, found, "Expected descriptor %d not found", expectedNAT.DescriptorID)
				}
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATStaticService_Update(t *testing.T) {
	tests := []struct {
		name        string
		nat         NATStatic
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update - add new entry",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
					{
						InsideLocal:   "192.168.1.101",
						OutsideGlobal: "203.0.113.101",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Get current state
				currentOutput := `nat descriptor type 1 static
nat descriptor static 1 203.0.113.100=192.168.1.100
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*1"`).
					Return([]byte(currentOutput), nil)

				// Add new entry using RunBatch
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					for _, cmd := range cmds {
						if cmd == "nat descriptor static 1 203.0.113.101=192.168.1.101" {
							return true
						}
					}
					return false
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful update - remove entry",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Get current state with two entries
				currentOutput := `nat descriptor type 1 static
nat descriptor static 1 203.0.113.100=192.168.1.100
nat descriptor static 1 203.0.113.101=192.168.1.101
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*1"`).
					Return([]byte(currentOutput), nil)

				// Delete old entry using RunBatch
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					for _, cmd := range cmds {
						if cmd == "no nat descriptor static 1 203.0.113.101=192.168.1.101" {
							return true
						}
					}
					return false
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error",
			nat: NATStatic{
				DescriptorID: 0, // Invalid
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid NAT static",
		},
		{
			name: "Get current state error",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor.*1"`).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "failed to get current NAT static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATStaticService{executor: mockExecutor}
			err := service.Update(context.Background(), tt.nat)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATStaticService_ContextCancellation(t *testing.T) {
	t.Run("Create with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		service := &NATStaticService{executor: new(MockExecutor)}
		err := service.Create(ctx, NATStatic{
			DescriptorID: 1,
			Entries: []NATStaticEntry{
				{InsideLocal: "192.168.1.100", OutsideGlobal: "203.0.113.100"},
			},
		})

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Get with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		service := &NATStaticService{executor: new(MockExecutor)}
		_, err := service.Get(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Delete with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		service := &NATStaticService{executor: new(MockExecutor)}
		err := service.Delete(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("List with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		service := &NATStaticService{executor: new(MockExecutor)}
		_, err := service.List(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Update with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		service := &NATStaticService{executor: new(MockExecutor)}
		err := service.Update(ctx, NATStatic{
			DescriptorID: 1,
			Entries: []NATStaticEntry{
				{InsideLocal: "192.168.1.100", OutsideGlobal: "203.0.113.100"},
			},
		})

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestNATStaticService_UsesRunBatch(t *testing.T) {
	t.Run("Create uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &NATStaticService{executor: mockExecutor}
		err := service.Create(context.Background(), NATStatic{
			DescriptorID: 1,
			Entries: []NATStaticEntry{
				{InsideLocal: "192.168.1.100", OutsideGlobal: "203.0.113.100"},
				{InsideLocal: "192.168.1.101", OutsideGlobal: "203.0.113.101"},
			},
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasType := false
		hasEntry1 := false
		hasEntry2 := false
		for _, cmd := range capturedCommands {
			if cmd == "nat descriptor type 1 static" {
				hasType = true
			}
			if cmd == "nat descriptor static 1 203.0.113.100=192.168.1.100" {
				hasEntry1 = true
			}
			if cmd == "nat descriptor static 1 203.0.113.101=192.168.1.101" {
				hasEntry2 = true
			}
		}
		assert.True(t, hasType, "Expected type command to be included")
		assert.True(t, hasEntry1, "Expected first entry command to be included")
		assert.True(t, hasEntry2, "Expected second entry command to be included")
	})

	t.Run("Delete uses RunBatch", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		mockExecutor.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
			Return([]byte(""), nil)

		service := &NATStaticService{executor: mockExecutor}
		err := service.Delete(context.Background(), 1)

		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})
}

func TestNATStaticService_EntriesEqual(t *testing.T) {
	service := &NATStaticService{}

	tests := []struct {
		name     string
		a        NATStaticEntry
		b        NATStaticEntry
		expected bool
	}{
		{
			name: "Equal 1:1 NAT entries",
			a: NATStaticEntry{
				InsideLocal:   "192.168.1.100",
				OutsideGlobal: "203.0.113.100",
			},
			b: NATStaticEntry{
				InsideLocal:   "192.168.1.100",
				OutsideGlobal: "203.0.113.100",
			},
			expected: true,
		},
		{
			name: "Equal port-based NAT entries",
			a: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8080),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "tcp",
			},
			b: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8080),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "TCP", // Case insensitive
			},
			expected: true,
		},
		{
			name: "Different inside local",
			a: NATStaticEntry{
				InsideLocal:   "192.168.1.100",
				OutsideGlobal: "203.0.113.100",
			},
			b: NATStaticEntry{
				InsideLocal:   "192.168.1.101",
				OutsideGlobal: "203.0.113.100",
			},
			expected: false,
		},
		{
			name: "Different outside global",
			a: NATStaticEntry{
				InsideLocal:   "192.168.1.100",
				OutsideGlobal: "203.0.113.100",
			},
			b: NATStaticEntry{
				InsideLocal:   "192.168.1.100",
				OutsideGlobal: "203.0.113.101",
			},
			expected: false,
		},
		{
			name: "Different ports",
			a: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8080),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "tcp",
			},
			b: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8081),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "tcp",
			},
			expected: false,
		},
		{
			name: "Different protocols",
			a: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8080),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "tcp",
			},
			b: NATStaticEntry{
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   staticIntPtr(8080),
				OutsideGlobal:     "203.0.113.100",
				OutsideGlobalPort: staticIntPtr(80),
				Protocol:          "udp",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.entriesEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
