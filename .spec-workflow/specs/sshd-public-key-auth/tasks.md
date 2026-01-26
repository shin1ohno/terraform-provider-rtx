# Tasks Document: SSHD Public Key Authentication

## Provider SSH Public Key Authentication

- [ ] 1. Extend Config struct with SSH key authentication fields
  - File: internal/client/interfaces.go
  - Add `PrivateKey`, `PrivateKeyFile`, `PrivateKeyPassphrase` fields to Config struct
  - Purpose: Data structure for SSH key authentication configuration
  - _Leverage: existing Config struct at line 917_
  - _Requirements: 1.1, 1.2, 1.3_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in SSH authentication | Task: Add three new fields to the Config struct in internal/client/interfaces.go: PrivateKey (string), PrivateKeyFile (string), PrivateKeyPassphrase (string). Add appropriate comments describing each field. | Restrictions: Do not modify other structs in the file, maintain consistent formatting | Success: Config struct has new fields, code compiles without errors | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 2. Implement SSH key authentication in ssh_dialer.go
  - File: internal/client/ssh_dialer.go
  - Add `buildAuthMethods()` function to construct auth methods list
  - Add `loadPrivateKey()` function to parse private key with optional passphrase
  - Add `trySSHAgent()` function to attempt SSH agent authentication
  - Modify `Dial()` to use buildAuthMethods instead of hardcoded password auth
  - Purpose: Enable provider to authenticate using SSH keys
  - _Leverage: existing sshDialer struct and Dial method, golang.org/x/crypto/ssh/agent_
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.6_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in SSH and crypto libraries | Task: Implement SSH public key authentication in internal/client/ssh_dialer.go. Create buildAuthMethods() that returns []ssh.AuthMethod with priority: 1) SSH agent (if no explicit key), 2) explicit private key, 3) password fallback. Create loadPrivateKey() using ssh.ParsePrivateKey and ssh.ParsePrivateKeyWithPassphrase. Create trySSHAgent() using agent.NewClient. Modify Dial() to use buildAuthMethods(). | Restrictions: Handle errors gracefully, log authentication attempts with Zerolog, do not remove existing password auth capability | Success: Provider can authenticate with private key file, private key content, SSH agent, or password. Unit tests pass. | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 3. Add provider schema attributes for SSH key authentication
  - File: internal/provider/provider.go
  - Add `private_key`, `private_key_file`, `private_key_passphrase` schema attributes
  - Pass values to Config struct in configure function
  - Purpose: Expose SSH key authentication options to Terraform users
  - _Leverage: existing provider schema and configure function_
  - _Requirements: 1.1, 1.2, 1.3, 1.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add three new schema attributes to provider.go: private_key (TypeString, Optional, Sensitive, env RTX_PRIVATE_KEY), private_key_file (TypeString, Optional, ConflictsWith private_key, env RTX_PRIVATE_KEY_FILE), private_key_passphrase (TypeString, Optional, Sensitive, env RTX_PRIVATE_KEY_PASSPHRASE). Update the configure function to read these values and set them on the Config struct. | Restrictions: Follow existing schema patterns, maintain backward compatibility with password auth | Success: Provider accepts new attributes, values are passed to client Config | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

## rtx_sshd Auth Method Extension

- [ ] 4. Extend SSHDConfig struct with AuthMethod field
  - File: internal/client/interfaces.go
  - Add `AuthMethod` field (string) to SSHDConfig struct
  - Purpose: Store authentication method configuration
  - _Leverage: existing SSHDConfig struct_
  - _Requirements: 4.1_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add AuthMethod string field to SSHDConfig struct in internal/client/interfaces.go. Add comment describing valid values: "password", "publickey", "any" (default). | Restrictions: Do not modify other fields | Success: SSHDConfig has AuthMethod field, code compiles | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 5. Add auth method commands to parsers
  - File: internal/rtx/parsers/service.go
  - Add `BuildSSHDAuthMethodCommand()` function
  - Add `BuildDeleteSSHDAuthMethodCommand()` function
  - Update `ParseSSHDConfig()` to extract auth method from output
  - Purpose: Generate and parse RTX auth method commands
  - _Leverage: existing SSHD command builders and parsers in service.go_
  - _Requirements: 4.1, 4.2, 4.3, 4.4_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in CLI parsing | Task: Add SSHD auth method support to internal/rtx/parsers/service.go. Create BuildSSHDAuthMethodCommand(method string) that returns "sshd auth method <method>" for "password"/"publickey", or "no sshd auth method" for "any"/empty. Create BuildDeleteSSHDAuthMethodCommand() returning "no sshd auth method". Update ParseSSHDConfig() to detect "sshd auth method" line and populate AuthMethod field. | Restrictions: Follow existing parser patterns, handle missing auth method gracefully (default to "any") | Success: Commands generated correctly, parser extracts auth method from config output | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 6. Update ServiceManager with auth method support
  - File: internal/client/service_manager.go
  - Update `ConfigureSSHD()` to set auth method
  - Update `UpdateSSHD()` to handle auth method changes
  - Purpose: Execute auth method commands on router
  - _Leverage: existing SSHD methods in ServiceManager_
  - _Requirements: 4.1, 4.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update SSHD methods in internal/client/service_manager.go. In ConfigureSSHD(): if config.AuthMethod is set and not "any", execute auth method command. In UpdateSSHD(): compare current vs new auth method, execute command if changed. | Restrictions: Maintain existing host and service logic, only add auth method handling | Success: SSHD configure/update sets auth method correctly, save is called after changes | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 7. Add auth_method attribute to rtx_sshd resource
  - File: internal/provider/resource_rtx_sshd.go
  - Add `auth_method` schema attribute (TypeString, Optional, Default "any")
  - Add validation for allowed values: "password", "publickey", "any"
  - Update buildSSHDConfigFromResourceData to include auth method
  - Update Read to set auth_method from config
  - Purpose: Expose auth method configuration to Terraform users
  - _Leverage: existing rtx_sshd resource schema and CRUD functions_
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add auth_method attribute to rtx_sshd resource in internal/provider/resource_rtx_sshd.go. Schema: TypeString, Optional, Default "any", ValidateFunc for "password"/"publickey"/"any". Update buildSSHDConfigFromResourceData to set AuthMethod. Update Read to call d.Set("auth_method", config.AuthMethod). | Restrictions: Maintain existing enabled/hosts/host_key attributes, follow existing patterns | Success: rtx_sshd accepts auth_method, value is persisted and read back correctly | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

## rtx_sshd_host_key Resource

- [ ] 8. Add host key service methods to ServiceManager
  - File: internal/client/service_manager.go
  - Add `GetSSHDHostKey()` method to check if host key exists and get fingerprint
  - Add `GenerateSSHDHostKey()` method to generate new host key
  - Purpose: Service layer for host key operations
  - _Leverage: existing ServiceManager patterns, parsers for command building_
  - _Requirements: 2.1, 2.2, 2.3_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add host key methods to ServiceManager in internal/client/service_manager.go. Create GetSSHDHostKey(ctx) (*SSHHostKeyInfo, error) that executes "show status sshd" and parses fingerprint/algorithm. Create GenerateSSHDHostKey(ctx) error that executes "sshd host key generate" and saves config. Define SSHHostKeyInfo struct in interfaces.go with Fingerprint and Algorithm fields. | Restrictions: Follow existing ServiceManager patterns, use Zerolog for logging | Success: Can check host key status and generate new key | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 9. Add host key parser functions
  - File: internal/rtx/parsers/service.go
  - Add `BuildShowSSHDStatusCommand()` function
  - Add `BuildSSHDHostKeyGenerateCommand()` function
  - Add `ParseSSHDHostKeyInfo()` function to extract fingerprint from show status output
  - Purpose: Parse RTX output for host key information
  - _Leverage: existing parser patterns in service.go_
  - _Requirements: 2.1, 2.4_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in CLI parsing | Task: Add host key parsers to internal/rtx/parsers/service.go. Create BuildShowSSHDStatusCommand() returning "show status sshd". Create BuildSSHDHostKeyGenerateCommand() returning "sshd host key generate". Create ParseSSHDHostKeyInfo(output string) (*SSHHostKeyInfo, error) that extracts fingerprint and algorithm from "show status sshd" output using regex. Return empty fingerprint if no key exists. | Restrictions: Handle missing host key gracefully, follow existing parser patterns | Success: Parser correctly extracts host key info from RTX output | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 10. Create rtx_sshd_host_key resource
  - File: internal/provider/resource_rtx_sshd_host_key.go
  - Create new resource with fingerprint (Computed) and algorithm (Computed) attributes
  - Implement Create: check if key exists, generate only if not, read key info
  - Implement Read: get current host key info
  - Implement Delete: no-op (host keys persist)
  - Implement Import: read existing host key info
  - Register resource in provider.go ResourcesMap
  - Purpose: Terraform resource for managing SSH host key
  - _Leverage: existing singleton resource patterns (rtx_sshd), ServiceManager host key methods_
  - _Requirements: 2.1, 2.2, 2.3, 2.4_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create rtx_sshd_host_key resource in internal/provider/resource_rtx_sshd_host_key.go. Schema: fingerprint (TypeString, Computed), algorithm (TypeString, Computed). Create: check GetSSHDHostKey(), if no fingerprint then GenerateSSHDHostKey(), set state from host key info, ID="sshd_host_key". Read: GetSSHDHostKey(), update state. Delete: return nil (no-op). Import: GetSSHDHostKey(), set state. Register in provider.go. | Restrictions: Never regenerate existing host key, follow singleton resource patterns from rtx_sshd | Success: Resource creates key only if missing, reads existing key info, delete is no-op | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

## rtx_sshd_authorized_keys Resource

- [ ] 11. Add authorized keys service methods to ServiceManager
  - File: internal/client/service_manager.go
  - Add `GetSSHDAuthorizedKeys()` method to list keys for a user
  - Add `SetSSHDAuthorizedKeys()` method to replace all keys for a user
  - Add `DeleteSSHDAuthorizedKeys()` method to remove all keys for a user
  - Purpose: Service layer for authorized keys operations
  - _Leverage: existing ServiceManager patterns, executor for interactive commands_
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in interactive CLI | Task: Add authorized keys methods to ServiceManager in internal/client/service_manager.go. Create GetSSHDAuthorizedKeys(ctx, username) ([]SSHAuthorizedKey, error) using "show sshd authorized-keys <user>". Create SetSSHDAuthorizedKeys(ctx, username, keys []string) error that deletes existing keys then imports each new key using "import sshd authorized-keys <user>" with key content as input. Create DeleteSSHDAuthorizedKeys(ctx, username) error using "delete /ssh/authorized_keys/<user>". Define SSHAuthorizedKey struct in interfaces.go. | Restrictions: Handle interactive import command correctly, save config after changes | Success: Can list, set, and delete authorized keys for a user | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 12. Add authorized keys parser functions
  - File: internal/rtx/parsers/service.go
  - Add `BuildShowSSHDAuthorizedKeysCommand()` function
  - Add `BuildImportSSHDAuthorizedKeysCommand()` function
  - Add `BuildDeleteSSHDAuthorizedKeysCommand()` function
  - Add `ParseSSHDAuthorizedKeys()` function to extract key list from output
  - Purpose: Parse RTX output for authorized keys
  - _Leverage: existing parser patterns in service.go_
  - _Requirements: 3.1, 3.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in CLI parsing | Task: Add authorized keys parsers to internal/rtx/parsers/service.go. Create BuildShowSSHDAuthorizedKeysCommand(username) returning "show sshd authorized-keys <user>". Create BuildImportSSHDAuthorizedKeysCommand(username) returning "import sshd authorized-keys <user>". Create BuildDeleteSSHDAuthorizedKeysCommand(username) returning "delete /ssh/authorized_keys/<user>". Create ParseSSHDAuthorizedKeys(output) ([]SSHAuthorizedKey, error) to parse fingerprint list from show output. | Restrictions: Follow existing parser patterns, handle empty key list gracefully | Success: Parser correctly extracts authorized key fingerprints from RTX output | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 13. Create rtx_sshd_authorized_keys resource
  - File: internal/provider/resource_rtx_sshd_authorized_keys.go
  - Create new resource with username (Required, ForceNew) and keys (Required, List) attributes
  - Implement Create: register all keys for user
  - Implement Read: get current keys, compare fingerprints
  - Implement Update: delete all keys, re-register desired keys
  - Implement Delete: delete all keys for user
  - Implement Import: read existing keys
  - Register resource in provider.go ResourcesMap
  - Purpose: Terraform resource for managing SSH authorized keys per user
  - _Leverage: existing resource patterns, ServiceManager authorized keys methods_
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create rtx_sshd_authorized_keys resource in internal/provider/resource_rtx_sshd_authorized_keys.go. Schema: username (TypeString, Required, ForceNew), keys (TypeList of TypeString, Required). Create: SetSSHDAuthorizedKeys() for all keys, ID=username. Read: GetSSHDAuthorizedKeys(), set state. Update: SetSSHDAuthorizedKeys() (which deletes and re-registers). Delete: DeleteSSHDAuthorizedKeys(). Import: ID is username, GetSSHDAuthorizedKeys(), set state. Register in provider.go. | Restrictions: Keys list comparison should be order-independent, follow resource patterns | Success: Resource manages authorized keys, update replaces all keys correctly | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

## Testing

- [ ] 14. Unit tests for SSH key authentication
  - File: internal/client/ssh_dialer_test.go
  - Test loadPrivateKey with various key formats (RSA, ED25519, encrypted)
  - Test buildAuthMethods priority order
  - Test trySSHAgent with mocked socket
  - Purpose: Ensure SSH key authentication works correctly
  - _Leverage: existing ssh_dialer_test.go patterns_
  - _Requirements: 1.1, 1.2, 1.3, 1.6_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with testing expertise | Task: Add unit tests to internal/client/ssh_dialer_test.go for SSH key authentication. Test loadPrivateKey with RSA key, ED25519 key, encrypted key with passphrase, invalid key. Test buildAuthMethods returns correct priority order. Use test fixtures for sample keys. | Restrictions: Do not require actual SSH connections, mock external dependencies | Success: All key formats are tested, auth method priority is verified | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 15. Unit tests for SSHD resources
  - File: internal/provider/resource_rtx_sshd_test.go, resource_rtx_sshd_host_key_test.go, resource_rtx_sshd_authorized_keys_test.go
  - Test auth_method attribute in rtx_sshd
  - Test host key idempotent behavior
  - Test authorized keys update logic
  - Purpose: Ensure resource logic works correctly
  - _Leverage: existing resource test patterns_
  - _Requirements: 2.1, 3.2, 4.1_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with testing expertise | Task: Add unit tests for SSHD resources. In resource_rtx_sshd_test.go: test auth_method validation, test buildSSHDConfigFromResourceData includes auth_method. Create resource_rtx_sshd_host_key_test.go: test Create with existing key (no regeneration), test Create without key (generates). Create resource_rtx_sshd_authorized_keys_test.go: test key list comparison, test Update triggers delete+re-register. | Restrictions: Use mocked ServiceManager, follow existing test patterns | Success: Resource logic is tested, edge cases covered | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [ ] 16. Acceptance tests with real RTX router
  - File: internal/provider/resource_rtx_sshd_acc_test.go, resource_rtx_sshd_host_key_acc_test.go, resource_rtx_sshd_authorized_keys_acc_test.go
  - Test full CRUD cycle for each resource
  - Test provider authentication with SSH key
  - Purpose: Verify end-to-end functionality with real router
  - _Leverage: existing acceptance test patterns, TF_ACC environment variable_
  - _Requirements: All_
  - _Prompt: Implement the task for spec sshd-public-key-auth, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer with Terraform testing expertise | Task: Create acceptance tests for SSHD resources. Test rtx_sshd with auth_method changes. Test rtx_sshd_host_key create/import on existing server. Test rtx_sshd_authorized_keys full lifecycle with multiple keys. All tests require TF_ACC=1 and real RTX router access. | Restrictions: Use build tags for acceptance tests, clean up resources after test | Success: All resources work correctly with real RTX router | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._
