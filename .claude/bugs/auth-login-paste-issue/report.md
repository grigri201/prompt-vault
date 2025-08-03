# Bug Report

## Bug Summary
The `pv auth login` command cannot accept pasted tokens properly, making it difficult for users to input their GitHub Personal Access Token. Additionally, the command does not support multi-byte character input (like Chinese characters).

## Bug Details

### Expected Behavior
- Users should be able to paste their entire GitHub Personal Access Token when prompted
- The command should support multi-byte character input (including Chinese)
- The pasted token should be accepted and processed correctly

### Actual Behavior  
- When pasting a token, only partial input is accepted or the paste operation fails
- Multi-byte characters (like Chinese) are not properly handled
- Users must manually type the token character by character, which is error-prone

### Steps to Reproduce
1. Run `pv auth login`
2. When prompted "Enter your GitHub Personal Access Token:"
3. Copy a GitHub Personal Access Token to clipboard
4. Attempt to paste the token (Ctrl+V on Windows/Linux, Cmd+V on Mac)
5. Observe that the full token is not pasted or accepted properly

### Environment
- **Version**: Current main branch
- **Platform**: Cross-platform issue (affects Windows, macOS, Linux)
- **Configuration**: Terminal with masked input mode (term.ReadPassword)

## Impact Assessment

### Severity
- [x] High - Major functionality broken

### Affected Users
All users attempting to authenticate with GitHub, particularly:
- Users with long Personal Access Tokens
- Users in regions using multi-byte character sets
- Users who rely on password managers

### Affected Features
- Authentication flow
- Initial setup experience
- Token management

## Additional Context

### Error Messages
```
No specific error messages - the issue is with input handling
```

### Screenshots/Media
N/A - Input masking prevents visual capture

### Related Issues
- Input handling in terminal applications
- Multi-byte character support in Go terminal packages

## Initial Analysis

### Suspected Root Cause
The `term.ReadPassword()` function in `cmd/auth_login.go` (line 77) may have limitations with:
1. Buffer size for pasted input
2. Multi-byte character handling
3. Terminal input mode settings that interfere with paste operations

### Affected Components
- `/cmd/auth_login.go` - specifically the `readToken()` function
- Terminal input handling using `golang.org/x/term` package
- Fallback to `bufio.NewReader` which may also have similar limitations