package echo

// TrustedOpenEchoExtensionOrigin pins OpenEcho's published extension identity.
// An unrelated extension must not be able to invoke local management operations.
// The manifest key is public. This origin does not grant MCP access.
const TrustedOpenEchoExtensionOrigin = "chrome-extension://khdciakflcmilakmaiejamcgmggnkego"
