package parsers

import "strings"

// dewrapConsoleLines reconstructs logical configuration lines from RTX
// `show config` output, undoing the router's terminal line-wrapping.
//
// RTX wraps long lines at the console width (default 80 columns) when writing to
// the terminal, regardless of the SSH PTY width the provider requests (512 wide).
// The wrap inserts a bare CRLF mid-line with no continuation marker or
// indentation. The break can land mid-token ("...classless_sta" + "tic_route=...")
// or at a space ("...edns=on " + "aaaa ."), so any line long enough to wrap loses
// data when the parser matches line-by-line — e.g. rtx_dhcp_scope option 121 and a
// rtx_dns_server IPv6 `dns server select` both vanished on read-back, producing
// "Provider produced inconsistent result after apply".
//
// Reassembly: a genuine logical line starts with a known keyword (isLogicalStart);
// any other non-empty line is a wrap continuation and is appended to the preceding
// logical line. Skipped (never joined): blank lines, `#` comment lines, the
// trailing interactive prompt ("[RTX1210] >"), and pre-output noise (the command
// echo, "Searching ...") before the first logical line. Joining the prompt or a
// comment onto a value was the bug that made the first de-wrap attempt ineffective.
//
// Join rule: continuations are concatenated with no separator so a mid-token break
// rejoins exactly ("classless_sta"+"tic_route="). When the continuation begins with
// "=" the break fell right before the "=" of a "key=value" token (e.g. "edns"+"=on")
// and RTX may have padded the first line to the column width, so its trailing
// whitespace is stripped first. An at-space break keeps the first line's trailing
// space, which is the real token separator ("...edns=on " + "aaaa .").
func dewrapConsoleLines(raw string, isLogicalStart func(trimmed string) bool) []string {
	// Normalize line endings (RTX emits CRLF; tolerate lone CR too).
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	physical := strings.Split(raw, "\n")
	logical := make([]string, 0, len(physical))
	for _, line := range physical {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			// blank line — skip
		case strings.HasPrefix(trimmed, "#"):
			// comment line — not config, not a wrap continuation
		case isLogicalStart(trimmed):
			logical = append(logical, line)
		case strings.HasSuffix(trimmed, ">") || strings.HasSuffix(trimmed, "#"):
			// RTX interactive prompt (e.g. "[RTX1210] >" / "# ") — config values
			// never end in '>' or '#'.
		case len(logical) == 0:
			// pre-output noise before the first logical line (command echo,
			// "Searching ...") — skip
		default:
			prev := logical[len(logical)-1]
			if strings.HasPrefix(trimmed, "=") {
				prev = strings.TrimRight(prev, " \t")
			}
			logical[len(logical)-1] = prev + line // wrap continuation
		}
	}
	return logical
}
