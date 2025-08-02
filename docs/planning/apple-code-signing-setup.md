# Apple Code Signing Setup for Plonk

This document outlines the complete process for setting up Apple code signing and notarization for plonk binaries.

**Status**: ✅ Completed and tested with v0.9.5 release

**Implementation**: Using GoReleaser's native notarize configuration instead of Quill for simpler setup.

## Prerequisites

- [x] Apple Developer Account ($99/year) - Already created
- [ ] Access to a macOS machine (for initial certificate creation only)
- [ ] GitHub repository admin access (for adding secrets)

## Step 1: Create Developer ID Application Certificate

### Option A: Using Xcode (Recommended)

1. Open Xcode on macOS
2. Go to **Xcode → Settings → Accounts** (or Preferences → Accounts on older versions)
3. Add your Apple ID if not already added
4. Select your team and click **Manage Certificates**
5. Click the **+** button and choose **Developer ID Application**
6. Xcode will create the certificate and install it in your Keychain

### Option B: Using Apple Developer Portal

1. Visit https://developer.apple.com/account/resources/certificates/list
2. Click the **+** button to create a new certificate
3. Select **Developer ID Application** under Software
4. You'll need to create a Certificate Signing Request (CSR):
   - Open **Keychain Access** on macOS
   - Menu: **Keychain Access → Certificate Assistant → Request a Certificate from a Certificate Authority**
   - Enter your email and name (must match your developer account)
   - Select **Saved to disk**
   - Save the CSR file
5. Upload the CSR file to the developer portal
6. Download the generated certificate
7. Double-click the downloaded certificate to install it in Keychain

## Step 2: Export Certificate as P12

1. Open **Keychain Access**
2. In the **login** keychain, find your certificate:
   - Look for "Developer ID Application: [Your Name] (TeamID)"
   - It should have a disclosure triangle showing the private key
3. Right-click on the certificate (not the private key)
4. Select **Export "Developer ID Application..."**
5. Save as `plonk-signing.p12`
6. Set a strong password when prompted - **save this password securely**
7. The export should include both the certificate and private key

## Step 3: Create App Store Connect API Key for Notarization

Apple requires an API key for notarization (not an app-specific password):

1. Go to https://appstoreconnect.apple.com/
2. Sign in with your Apple Developer account
3. Navigate to **Users and Access**
4. Click on the **Keys** tab under **Integrations**
5. Click the **+** button to generate a new API key
6. Set the following:
   - **Name**: "plonk-notarization" or similar
   - **Access**: "Developer" role (minimum required for notarization)
7. Click **Generate**
8. **IMPORTANT**: Download the `.p8` private key file immediately - you can only download it once!
9. Note down the following information:
   - **Key ID**: Shown in the list (like "ABC123DEFG")
   - **Issuer ID**: Found at the top of the Keys page (UUID format)
10. Store the `.p8` file securely

## Step 4: Find Your Team ID

1. Go to https://developer.apple.com/account
2. In the membership section, find your **Team ID** (10 characters, like `ABCD1234EF`)
3. Note this down - you'll need it for notarization

## Step 5: Prepare Secrets for GitHub

Convert your P12 certificate and P8 key to base64 for GitHub secrets:

```bash
# Convert P12 certificate
base64 -i plonk-signing.p12 > plonk-signing-base64.txt

# Convert P8 API key
base64 -i AuthKey_XXXXXXXXXX.p8 > notary-key-base64.txt

# Copy P12 to clipboard (macOS)
cat plonk-signing-base64.txt | pbcopy

# After adding P12, copy P8 to clipboard
cat notary-key-base64.txt | pbcopy
```

## Step 6: Add GitHub Repository Secrets

Go to your plonk repository: **Settings → Secrets and variables → Actions**

Add these repository secrets:

| Secret Name | Value |
|------------|-------|
| `QUILL_SIGN_P12` | The base64-encoded P12 certificate (from step 5) |
| `QUILL_SIGN_PASSWORD` | The password you set when exporting the P12 |
| `QUILL_NOTARY_KEY` | The base64-encoded P8 API key (from step 5) |
| `QUILL_NOTARY_KEY_ID` | Your Key ID from App Store Connect (like "ABC123DEFG") |
| `QUILL_NOTARY_ISSUER` | Your Issuer ID from App Store Connect (UUID format) |

## Step 7: Update GoReleaser Configuration

Update `.goreleaser.yaml` to use GoReleaser's native notarize configuration:

```yaml
# Add after the archives section
notarize:
  macos:
    - enabled: '{{ isEnvSet "QUILL_SIGN_P12" }}'
      ids:
        - plonk  # Your build ID
      sign:
        certificate: "{{.Env.QUILL_SIGN_P12}}"
        password: "{{.Env.QUILL_SIGN_PASSWORD}}"
      notarize:
        issuer_id: "{{.Env.QUILL_NOTARY_ISSUER}}"
        key_id: "{{.Env.QUILL_NOTARY_KEY_ID}}"
        key: "{{.Env.QUILL_NOTARY_KEY}}"
        wait: true
        timeout: 20m

# The homebrew_casks section remains the same
homebrew_casks:
  - name: plonk
    repository:
      owner: richhaase
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/richhaase/plonk"
    description: "The unified package and dotfile manager for developers who tinker"
    license: "MIT"
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
    commit_msg_template: "Cask update for {{ .ProjectName }} version {{ .Tag }}"
    # No hooks needed - signing handles quarantine
```

## Step 8: Update GitHub Actions Workflow

Update `.github/workflows/release.yml` to pass the signing environment variables:

```yaml
- name: Checkout
  uses: actions/checkout@v4
  with:
    fetch-depth: 0

- name: Setup Go environment
  uses: ./.github/actions/setup-go-env
  with:
    go-version: 'stable'

- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v6
  with:
    distribution: goreleaser
    version: v2.11.0
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
    QUILL_SIGN_P12: ${{ secrets.QUILL_SIGN_P12 }}
    QUILL_SIGN_PASSWORD: ${{ secrets.QUILL_SIGN_PASSWORD }}
    QUILL_NOTARY_KEY: ${{ secrets.QUILL_NOTARY_KEY }}
    QUILL_NOTARY_KEY_ID: ${{ secrets.QUILL_NOTARY_KEY_ID }}
    QUILL_NOTARY_ISSUER: ${{ secrets.QUILL_NOTARY_ISSUER }}
```

Note: GoReleaser handles the signing tools internally, so no need to install Quill separately.

## Step 9: Test Locally (Optional)

To test signing locally before pushing:

1. Install Quill:
   ```bash
   curl -sSfL https://raw.githubusercontent.com/anchore/quill/main/install.sh | sh -s -- -b /usr/local/bin
   ```

2. Test signing:
   ```bash
   export QUILL_SIGN_P12=/path/to/plonk-signing.p12
   export QUILL_SIGN_PASSWORD="your-p12-password"

   # Test sign only (faster)
   quill sign /path/to/plonk-binary

   # Test full sign and notarize (takes 5-10 minutes)
   export QUILL_NOTARY_KEY="your-app-specific-password"
   export QUILL_NOTARY_KEY_ID="your-apple-id@example.com"
   export QUILL_NOTARY_ISSUER="TEAMID"

   quill sign-and-notarize /path/to/plonk-binary
   ```

3. Verify signature:
   ```bash
   codesign -dvv /path/to/signed/plonk
   ```

## Step 10: Release Process

1. Commit all changes:
   ```bash
   git add .goreleaser.yaml .github/workflows/release.yml
   git commit -m "feat: add Apple code signing and notarization"
   git push
   ```

2. Create a test release:
   ```bash
   git tag v0.9.3
   git push origin v0.9.3
   ```

3. Monitor the GitHub Actions workflow
4. Check that binaries are signed and notarized
5. Test installation via Homebrew

## Troubleshooting

### Common Issues

1. **"Certificate not found"**
   - Ensure the P12 contains both certificate and private key
   - Check the certificate name matches exactly

2. **"Failed to notarize"**
   - Verify app-specific password is correct
   - Check Team ID is correct
   - Ensure Apple ID matches the certificate

3. **"Invalid signature"**
   - Certificate might be expired
   - P12 might be corrupted during base64 encoding

### Verification Commands

```bash
# Check certificate details
quill p12 describe /path/to/plonk-signing.p12

# Verify signed binary
codesign -dvv /path/to/signed/plonk

# Check notarization status
spctl -a -vvv -t install /path/to/signed/plonk
```

## Benefits of Code Signing

1. **No quarantine warnings** - Users can run plonk immediately
2. **Trust indicator** - Shows "Verified Developer" in security settings
3. **Notarization** - Apple scans for malware
4. **Professional distribution** - Expected for production tools

## Notes

- Notarization typically takes 5-10 minutes
- Certificates expire after 5 years (renewable)
- The same certificate can be used for all your projects
- Quill enables signing from Linux/Windows CI environments

## References

- [Quill Documentation](https://github.com/anchore/quill)
- [Apple Developer - Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [GoReleaser Signing Documentation](https://goreleaser.com/customization/sign/)
