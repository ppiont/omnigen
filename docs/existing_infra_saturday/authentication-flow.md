# Authentication Flow

> Cognito OAuth2/OIDC authentication with JWT validation and security controls

## Overview

OmniGen uses **AWS Cognito User Pool** for authentication, implementing the **OAuth 2.0 Authorization Code Grant** flow with **OIDC (OpenID Connect)** tokens. All API requests require valid JWT tokens validated against Cognito's public keys.

**Security Features:**
- OAuth2/OIDC standard compliance
- JWT with RS256 signature verification
- HttpOnly cookies for XSS protection (future)
- Rate limiting (100 requests/min per user)
- Quota enforcement (10 videos/month per user)
- HTTPS-only communication
- JWKS key rotation support

**Cognito Configuration:**
- User Pool ID: `us-east-1_{random}`
- Client ID: App client for frontend
- Hosted UI: `https://omnigen.auth.us-east-1.amazoncognito.com`
- Token Expiration: Access token (1 hour), Refresh token (30 days)

---

## User Registration Flow

New user sign-up via Cognito Hosted UI with email verification.

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant Cognito as Cognito Hosted UI
    participant Pool as Cognito User Pool
    participant Email as Email Service<br/>(AWS SES)

    User->>React: Click "Sign Up"
    React->>Cognito: Redirect to /oauth2/authorize<br/>?response_type=code<br/>&client_id={clientId}<br/>&redirect_uri=https://{cloudfront}/callback<br/>&scope=openid email profile

    Cognito->>User: Show sign-up form
    User->>Cognito: Enter email, password, name
    Cognito->>Pool: CreateUser<br/>{email, password, attributes}

    alt Password too weak
        Pool-->>Cognito: Error: Password policy violation
        Cognito-->>User: Show error (min 8 chars, uppercase, number, symbol)
    else Email already exists
        Pool-->>Cognito: Error: User already exists
        Cognito-->>User: Show error + "Forgot password?" link
    else Valid registration
        Pool->>Pool: Hash password (bcrypt)
        Pool->>Email: Send verification email<br/>Code: {6-digit code}
        Email-->>User: Email with verification link

        Pool-->>Cognito: User created (status: UNVERIFIED)
        Cognito-->>User: Show "Check your email" message

        User->>Email: Click verification link<br/>https://cognito...?code={code}
        Email->>Pool: ConfirmSignUp<br/>{email, code}
        Pool->>Pool: Update status: CONFIRMED
        Pool-->>Cognito: User verified

        Cognito->>Pool: InitiateAuth<br/>{email, password}
        Pool->>Pool: Validate credentials
        Pool-->>Cognito: Tokens: {access, id, refresh}

        Cognito->>React: Redirect to /callback<br/>?code={authorizationCode}
        React->>Cognito: POST /oauth2/token<br/>grant_type=authorization_code<br/>code={authorizationCode}<br/>redirect_uri=https://{cloudfront}/callback

        Cognito->>Pool: ValidateAuthCode
        Pool-->>Cognito: Tokens
        Cognito-->>React: JSON Response:<br/>{<br/>  "access_token": "...",<br/>  "id_token": "...",<br/>  "refresh_token": "...",<br/>  "expires_in": 3600<br/>}

        React->>React: Store tokens in localStorage<br/>Parse JWT claims (sub, email)
        React->>React: Redirect to /dashboard
    end
```

**Password Policy (Cognito):**
- Minimum length: 8 characters
- Require: Uppercase, lowercase, number, special character
- No common passwords (checked against breach database)

**Email Verification:**
- Code valid for: 24 hours
- Resend limit: 5 per hour (Cognito throttling)

---

## OAuth2 Login Flow

Existing user login via Cognito Hosted UI (Authorization Code Grant).

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant CF as CloudFront
    participant Cognito as Cognito Hosted UI
    participant Pool as Cognito User Pool

    User->>React: Click "Login"
    React->>Cognito: Redirect GET /oauth2/authorize<br/>?response_type=code<br/>&client_id={clientId}<br/>&redirect_uri=https://{cloudfront}/callback<br/>&scope=openid email profile<br/>&state={randomString}

    Note over Cognito: Cognito checks session cookie<br/>If exists, skip login form

    Cognito->>User: Show login form
    User->>Cognito: Enter email + password
    Cognito->>Pool: InitiateAuth<br/>AUTH_FLOW: USER_PASSWORD_AUTH<br/>{email, password}

    alt Invalid credentials
        Pool-->>Cognito: Error: Incorrect username or password
        Cognito-->>User: Show error + "Forgot password?"
    else User not confirmed
        Pool-->>Cognito: Error: User not confirmed
        Cognito-->>User: Show "Verify your email" + resend link
    else Valid credentials
        Pool->>Pool: Validate password hash
        Pool->>Pool: Generate authorization code<br/>Code: {randomString} (valid 5 min)
        Pool-->>Cognito: Authorization code

        Cognito->>React: 302 Redirect to:<br/>https://{cloudfront}/callback<br/>?code={authorizationCode}<br/>&state={originalState}

        React->>React: Validate state matches original

        React->>Cognito: POST /oauth2/token<br/>Content-Type: application/x-www-form-urlencoded<br/><br/>grant_type=authorization_code<br/>code={authorizationCode}<br/>client_id={clientId}<br/>redirect_uri=https://{cloudfront}/callback

        Cognito->>Pool: ValidateAuthCode<br/>{code, clientId, redirectUri}

        alt Code expired (>5 min)
            Pool-->>Cognito: Error: Invalid authorization code
            Cognito-->>React: 400 Bad Request
            React-->>User: Redirect to /login (retry)
        else Code valid
            Pool->>Pool: Generate JWT tokens<br/>Access token (RS256, 1 hour)<br/>ID token (RS256, 1 hour)<br/>Refresh token (opaque, 30 days)

            Pool-->>Cognito: Token set
            Cognito-->>React: 200 OK<br/>Content-Type: application/json<br/><br/>{<br/>  "access_token": "eyJraWQ...",<br/>  "id_token": "eyJraWQ...",<br/>  "refresh_token": "...",<br/>  "expires_in": 3600,<br/>  "token_type": "Bearer"<br/>}

            React->>React: Parse ID token (base64 decode)<br/>Extract claims:<br/>{<br/>  "sub": "uuid",<br/>  "email": "user@example.com",<br/>  "cognito:username": "user@example.com"<br/>}

            React->>React: Store in localStorage:<br/>- accessToken<br/>- idToken<br/>- refreshToken<br/>- expiresAt (now + 3600s)

            React->>CF: GET /api/v1/user/me<br/>Authorization: Bearer {accessToken}

            Note over CF,Pool: JWT validation flow (see below)

            CF-->>React: 200 OK<br/>{"userId": "...", "email": "...", "quota": {...}}

            React->>React: Set auth context<br/>Redirect to /dashboard
        end
    end
```

**OAuth2 Parameters:**
- `response_type=code`: Authorization Code Grant (most secure for SPAs)
- `client_id`: Public client (no client secret needed for PKCE)
- `redirect_uri`: Must match Cognito app client configuration
- `scope`: `openid email profile` (OIDC standard scopes)
- `state`: CSRF protection (random string validated on callback)

**Token Storage:**
- **Current:** localStorage (accessible to JavaScript, XSS risk)
- **Future:** HttpOnly cookies (immune to XSS, requires backend cookie management)

---

## JWT Validation Flow

Backend validates JWT on every API request using Cognito's public keys (JWKS).

```mermaid
sequenceDiagram
    participant React as React Frontend
    participant ALB as Application LB
    participant ECS as ECS API<br/>Go Gin Server
    participant JWKS as Cognito JWKS<br/>Public Keys
    participant DDB as DynamoDB<br/>Usage Table

    React->>ALB: GET /api/v1/jobs<br/>Authorization: Bearer eyJraWQiOiJhYmMxMjMi...
    ALB->>ECS: Forward request

    Note over ECS: Middleware: JWT Auth

    ECS->>ECS: Extract Bearer token<br/>Header: Authorization

    alt No token
        ECS-->>ALB: 401 Unauthorized<br/>WWW-Authenticate: Bearer realm="OmniGen API"
        ALB-->>React: 401 Unauthorized
    else Token present
        ECS->>ECS: Decode JWT header (base64)<br/>Extract "kid" (Key ID)

        alt JWKS not in cache
            ECS->>JWKS: GET /.well-known/jwks.json<br/>https://cognito-idp.us-east-1.amazonaws.com<br/>/{userPoolId}/.well-known/jwks.json
            JWKS-->>ECS: JSON:<br/>{<br/>  "keys": [<br/>    {<br/>      "kid": "abc123",<br/>      "kty": "RSA",<br/>      "n": "...",<br/>      "e": "AQAB"<br/>    }<br/>  ]<br/>}
            ECS->>ECS: Cache JWKS in-memory<br/>TTL: 1 hour
        else JWKS in cache
            ECS->>ECS: Use cached JWKS
        end

        ECS->>ECS: Find public key by "kid"

        alt Key not found
            ECS-->>ALB: 401 Unauthorized<br/>Error: Unknown key ID
            ALB-->>React: 401 Unauthorized
        else Key found
            ECS->>ECS: Verify JWT signature<br/>Algorithm: RS256<br/>Public key: RSA (from JWKS)

            alt Signature invalid
                ECS-->>ALB: 401 Unauthorized<br/>Error: Invalid token signature
                ALB-->>React: 401 Unauthorized
            else Signature valid
                ECS->>ECS: Decode JWT payload<br/>Extract claims:<br/>{<br/>  "sub": "uuid",<br/>  "email": "...",<br/>  "exp": 1702915200,<br/>  "iss": "https://cognito-idp..."<br/>}

                ECS->>ECS: Validate claims:<br/>1. iss = expected issuer<br/>2. aud/client_id = expected<br/>3. exp > now (not expired)<br/>4. nbf <= now (not before)

                alt Token expired
                    ECS-->>ALB: 401 Unauthorized<br/>Error: Token expired
                    ALB-->>React: 401 Unauthorized<br/>+ Trigger token refresh
                else Invalid issuer
                    ECS-->>ALB: 401 Unauthorized<br/>Error: Invalid issuer
                    ALB-->>React: 401 Unauthorized
                else All validations pass
                    ECS->>ECS: Set Gin context:<br/>c.Set("userId", claims.Sub)<br/>c.Set("email", claims.Email)

                    Note over ECS: Middleware: Rate Limit

                    ECS->>ECS: Check in-memory counter<br/>Key: userId<br/>Limit: 100 req/min

                    alt Rate limit exceeded
                        ECS-->>ALB: 429 Too Many Requests<br/>Retry-After: 60
                        ALB-->>React: 429 Too Many Requests
                    else Rate limit OK
                        Note over ECS: Middleware: Quota

                        ECS->>DDB: GetItem<br/>Table: omnigen-usage<br/>Key: userId

                        alt DynamoDB error
                            ECS-->>ALB: 503 Service Unavailable
                            ALB-->>React: 503 Service Unavailable
                        else Quota exceeded
                            ECS-->>ALB: 429 Too Many Requests<br/>Error: Monthly quota exceeded
                            ALB-->>React: 429 Too Many Requests
                        else All checks pass
                            ECS->>ECS: Execute handler:<br/>jobs.GetJobs(c)
                            ECS-->>ALB: 200 OK + JSON
                            ALB-->>React: 200 OK + JSON
                        end
                    end
                end
            end
        end
    end
```

**JWT Structure (Access Token):**

**Header:**
```json
{
  "kid": "abc123",
  "alg": "RS256"
}
```

**Payload:**
```json
{
  "sub": "e5b3f8d2-1234-5678-9abc-def012345678",
  "email": "user@example.com",
  "cognito:username": "user@example.com",
  "iss": "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_ABC123",
  "client_id": "7abcdefghijk1234567890",
  "origin_jti": "...",
  "token_use": "access",
  "scope": "openid email profile",
  "auth_time": 1702914000,
  "exp": 1702917600,
  "iat": 1702914000,
  "jti": "..."
}
```

**Signature:** RS256 (RSA SHA-256) verified with Cognito public key

**Performance:**
- JWKS fetch: ~100ms (first request only, then cached 1 hour)
- Signature verification: ~5ms (CPU-bound crypto operation)
- Claims validation: <1ms (simple comparisons)
- **Total:** ~5-10ms per request (after JWKS cached)

---

## Token Refresh Flow

Automatically refresh expired access tokens using long-lived refresh tokens.

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant Cognito as Cognito Token<br/>Endpoint
    participant Pool as Cognito User Pool

    Note over React: Access token expired<br/>(checked before API call:<br/>if expiresAt < now)

    React->>Cognito: POST /oauth2/token<br/>Content-Type: application/x-www-form-urlencoded<br/><br/>grant_type=refresh_token<br/>client_id={clientId}<br/>refresh_token={refreshToken}

    Cognito->>Pool: ValidateRefreshToken<br/>{refreshToken}

    alt Refresh token expired (>30 days)
        Pool-->>Cognito: Error: Refresh token expired
        Cognito-->>React: 400 Bad Request<br/>{"error": "invalid_grant"}
        React->>React: Clear localStorage<br/>Redirect to /login
    else Refresh token revoked
        Pool-->>Cognito: Error: Token has been revoked
        Cognito-->>React: 400 Bad Request
        React->>React: Redirect to /login
    else Refresh token valid
        Pool->>Pool: Generate new access token<br/>Generate new ID token<br/>(Refresh token unchanged)

        Pool-->>Cognito: New tokens
        Cognito-->>React: 200 OK<br/>{<br/>  "access_token": "eyJraWQ...",<br/>  "id_token": "eyJraWQ...",<br/>  "expires_in": 3600,<br/>  "token_type": "Bearer"<br/>}

        React->>React: Update localStorage:<br/>- accessToken (new)<br/>- idToken (new)<br/>- expiresAt (now + 3600s)<br/>- refreshToken (unchanged)

        React->>React: Retry failed API request<br/>with new access token

        Note over User: User session extended<br/>(transparent to user)
    end
```

**Refresh Token Characteristics:**
- **Lifetime:** 30 days (configurable in Cognito)
- **Format:** Opaque string (not JWT, cannot be decoded)
- **Rotation:** Cognito can rotate refresh tokens on use (disabled by default)
- **Revocation:** User logout or admin action revokes all refresh tokens

**Frontend Implementation:**
```javascript
// Axios interceptor for automatic token refresh
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const { data } = await axios.post('https://cognito.../oauth2/token', {
            grant_type: 'refresh_token',
            client_id: CLIENT_ID,
            refresh_token: refreshToken
          });
          localStorage.setItem('accessToken', data.access_token);
          localStorage.setItem('idToken', data.id_token);
          localStorage.setItem('expiresAt', Date.now() + data.expires_in * 1000);

          // Retry original request with new token
          error.config.headers.Authorization = `Bearer ${data.access_token}`;
          return axios.request(error.config);
        } catch (refreshError) {
          // Refresh failed, logout user
          localStorage.clear();
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);
```

---

## Password Reset Flow

Forgot password flow via Cognito with email verification code.

```mermaid
sequenceDiagram
    actor User
    participant React as React Frontend
    participant Cognito as Cognito Hosted UI
    participant Pool as Cognito User Pool
    participant Email as Email Service

    User->>React: Click "Forgot Password"
    React->>Cognito: Redirect to /forgotPassword

    Cognito->>User: Show "Enter your email" form
    User->>Cognito: Enter email
    Cognito->>Pool: ForgotPassword<br/>{email}

    alt Email not found
        Pool-->>Cognito: Success (don't reveal user existence)
        Cognito-->>User: "If email exists, code sent"
    else Email exists
        Pool->>Email: Send password reset code<br/>Code: {6-digit number}
        Email-->>User: Email with code (valid 1 hour)

        Pool-->>Cognito: Success
        Cognito-->>User: Show "Check your email"

        User->>Email: Copy code from email
        User->>Cognito: Enter code + new password

        Cognito->>Pool: ConfirmForgotPassword<br/>{email, code, newPassword}

        alt Code expired
            Pool-->>Cognito: Error: Code expired
            Cognito-->>User: Show error + "Resend code"
        else Code invalid
            Pool-->>Cognito: Error: Invalid code
            Cognito-->>User: Show error (3 attempts remaining)
        else Password too weak
            Pool-->>Cognito: Error: Password policy violation
            Cognito-->>User: Show password requirements
        else All valid
            Pool->>Pool: Hash new password (bcrypt)
            Pool->>Pool: Invalidate all refresh tokens<br/>(force re-login on all devices)
            Pool-->>Cognito: Password reset successful

            Cognito-->>User: Show "Password reset successful"<br/>Redirect to login

            User->>Cognito: Enter email + new password
            Note over Cognito,Pool: Standard OAuth2 login flow
        end
    end
```

**Security Measures:**
- **Rate Limiting:** 5 reset requests per hour per email (Cognito enforced)
- **Code Expiration:** 1 hour (configurable)
- **Code Attempts:** 3 attempts before code invalidation
- **User Enumeration Protection:** Always return success, even if email not found
- **Token Revocation:** All refresh tokens invalidated on password reset

---

## Security Architecture

Comprehensive view of authentication and authorization security layers.

```mermaid
flowchart TB
    User([End User])

    subgraph Frontend[\"React Frontend (CloudFront)\"]
        Login[Login Component]
        Storage[Token Storage<br/>localStorage]
        Interceptor[Axios Interceptor<br/>Auto-refresh]
    end

    subgraph Cognito[\"AWS Cognito\"]
        HostedUI[Hosted UI<br/>OAuth2 Flows]
        UserPool[User Pool<br/>Identity Store]
        JWKS_Endpoint[JWKS Endpoint<br/>Public Keys]
        TokenEndpoint[Token Endpoint<br/>/oauth2/token]
    end

    subgraph Backend[\"Go API (ECS Fargate)\"]
        Middleware[Middleware Stack]

        subgraph AuthStack[\"Authentication Layers\"]
            JWTMiddleware[JWT Validation<br/>RS256 Signature]
            RateLimit[Rate Limiter<br/>100 req/min]
            Quota[Quota Enforcement<br/>10 videos/month]
        end

        Handler[Route Handlers<br/>Business Logic]
    end

    subgraph Storage[\"Data Layer\"]
        DDB_Usage[DynamoDB Usage<br/>Quota Tracking]
        CloudWatch[CloudWatch Logs<br/>Audit Trail]
    end

    subgraph Security[\"Security Controls\"]
        HTTPS[TLS 1.2+<br/>All Communication]
        CORS_Policy[CORS Policy<br/>CloudFront Domain]
        CSP[Content Security Policy<br/>XSS Prevention]
        PasswordPolicy[Password Policy<br/>8+ chars, complexity]
    end

    User --> Login
    Login --> HostedUI
    HostedUI --> UserPool
    UserPool --> TokenEndpoint
    TokenEndpoint --> Storage
    Storage --> Interceptor
    Interceptor --> Middleware

    Middleware --> JWTMiddleware
    JWTMiddleware -.->|Fetch public keys| JWKS_Endpoint
    JWTMiddleware --> RateLimit
    RateLimit --> Quota
    Quota --> Handler

    Quota --> DDB_Usage
    Handler --> CloudWatch

    HTTPS -.->|Enforces| User
    HTTPS -.->|Enforces| Backend
    CORS_Policy -.->|Protects| Backend
    CSP -.->|Protects| Frontend
    PasswordPolicy -.->|Enforces| UserPool

    style AuthStack fill:#f8bbd0,stroke:#c2185b,stroke-width:2px
    style Security fill:#c8e6c9,stroke:#388e3c,stroke-width:2px
    style Cognito fill:#e1f5ff,stroke:#0288d1,stroke-width:2px
```

---

## Security Best Practices

### Current Implementation

| Security Control | Status | Details |
|-----------------|--------|---------|
| **HTTPS Only** | ✅ Enforced | CloudFront + ALB enforce TLS 1.2+ |
| **JWT Signature Verification** | ✅ Implemented | RS256 with Cognito public keys (JWKS) |
| **Token Expiration** | ✅ Enforced | Access token: 1 hour, Refresh: 30 days |
| **Rate Limiting** | ✅ Implemented | 100 requests/min per user (in-memory) |
| **Quota Enforcement** | ✅ Implemented | 10 videos/month per user (DynamoDB) |
| **CORS Policy** | ⚠️ Permissive | `Access-Control-Allow-Origin: *` (MVP only) |
| **HttpOnly Cookies** | ❌ Not Implemented | Tokens stored in localStorage (XSS risk) |
| **CSRF Protection** | ✅ Partial | OAuth2 `state` parameter validates redirects |
| **Password Policy** | ✅ Enforced | 8+ chars, uppercase, lowercase, number, symbol |
| **Audit Logging** | ✅ Implemented | CloudWatch Logs (all auth events) |

### Production Hardening (Future)

**High Priority:**
1. **HttpOnly Cookies for Tokens**
   - Move from localStorage to secure HttpOnly cookies
   - Backend manages cookie lifecycle (Set-Cookie header)
   - Prevents XSS token theft

2. **Restrict CORS to CloudFront Domain**
   ```javascript
   Access-Control-Allow-Origin: https://{cloudfront-domain}
   Access-Control-Allow-Credentials: true
   ```

3. **Implement PKCE (Proof Key for Code Exchange)**
   - Add `code_challenge` and `code_verifier` to OAuth2 flow
   - Protects against authorization code interception

4. **Add Content Security Policy (CSP)**
   ```http
   Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-{random}'
   ```

**Medium Priority:**
5. **Multi-Factor Authentication (MFA)**
   - Cognito supports TOTP (Google Authenticator) and SMS
   - Enforce for admin accounts

6. **Persistent Rate Limiting**
   - Move from in-memory to Redis/DynamoDB
   - Prevents rate limit bypass via ECS task restart

7. **IP-Based Geo-Blocking**
   - AWS WAF to block high-risk countries
   - Reduce credential stuffing attacks

8. **Advanced Threat Protection**
   - Cognito Advanced Security (risk-based auth)
   - Detects compromised credentials, unusual login patterns

---

## Authentication Metrics

### Cognito Limits (Free Tier)

| Metric | Free Tier | Paid Tier |
|--------|-----------|-----------|
| **Monthly Active Users (MAU)** | 50,000 free | $0.0055 per MAU |
| **Sign-up/Sign-in Requests** | Unlimited | Unlimited |
| **Token Refresh Requests** | Unlimited | Unlimited |
| **JWKS Endpoint** | Unlimited | Unlimited |
| **SMS MFA** | 50 free/month | $0.00645 per SMS |

**Current Usage:** ~10 users (well under free tier)

### Performance Metrics

| Operation | Latency | Cost |
|-----------|---------|------|
| **OAuth2 Login** | 500-1000ms | Free (Cognito) |
| **JWT Validation** | 5-10ms | $0 (CPU cost included in ECS) |
| **JWKS Fetch** | 100ms | Free (cached 1 hour) |
| **Token Refresh** | 200-300ms | Free (Cognito) |
| **Rate Limit Check** | <1ms | $0 (in-memory) |
| **Quota Check (DynamoDB)** | 5-10ms | $0.00005 per request |

---

## Troubleshooting

### Common Issues

**1. Token Expired Error**
```
Error: Token expired (exp claim)
Solution: Frontend should auto-refresh before expiration
Check: localStorage.getItem('expiresAt') < Date.now()
```

**2. Invalid Signature Error**
```
Error: Invalid token signature
Causes:
- Token tampered with
- JWKS not updated after key rotation
- Wrong user pool ID in backend config
Solution: Clear localStorage, re-login
```

**3. 429 Too Many Requests**
```
Error: Rate limit exceeded
Solution: Wait 60 seconds, implement exponential backoff
Check: CloudWatch Logs for excessive requests from userId
```

**4. CORS Error in Browser**
```
Error: CORS policy blocked
Causes:
- Origin not in allowed origins
- Preflight (OPTIONS) failed
Solution: Update ALB/CloudFront CORS headers
```

**5. Redirect URI Mismatch**
```
Error: redirect_uri doesn't match configured URI
Solution: Ensure exact match in Cognito app client settings
Example: https://d1234567890.cloudfront.net/callback
```

---

**Related Documentation:**
- [Architecture Overview](./architecture-overview.md) - System design with Cognito integration
- [Data Flow](./data-flow.md) - Request/response sequences with JWT validation
- [Backend Architecture](./backend-architecture.md) - Go middleware implementation
