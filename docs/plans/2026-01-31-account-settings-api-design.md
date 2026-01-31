# Account Settings API Design

## Overview

This document describes the API design for user account settings in RevieU Backend, supporting a social/content + e-commerce app.

## Route Structure

Split into two major route groups:
- `/auth` - Authentication (login, register, OAuth, verify)
- `/user` - User management (profile, settings, addresses) - requires JWT

### Complete Route Map

```
/auth                                   # Authentication (existing)
├── POST   /register
├── POST   /login
├── GET    /login/google
├── GET    /callback/google
├── GET    /verify
└── GET    /me

/user                                   # User Management (new, all require JWT)
├── /profile
│   ├── GET    /                        # Get profile
│   └── PATCH  /                        # Update profile
├── /security
│   ├── GET    /                        # Get security overview
│   ├── POST   /password                # Change password
│   ├── POST   /link/:provider          # Link OAuth
│   └── DELETE /link/:provider          # Unlink OAuth
├── /privacy
│   ├── GET    /                        # Get privacy settings
│   └── PATCH  /                        # Update privacy settings
├── /notifications
│   ├── GET    /                        # Get notification settings
│   └── PATCH  /                        # Update notification settings
├── /addresses
│   ├── GET    /                        # List addresses
│   ├── POST   /                        # Create address
│   ├── PATCH  /:id                     # Update address
│   ├── DELETE /:id                     # Delete address
│   └── POST   /:id/default             # Set as default
└── /account
    ├── POST   /export                  # Request data export
    └── DELETE /                        # Delete account
```

## Data Models

### Existing Models (unchanged)

- `User` - Core user table
- `UserAuth` - Authentication methods
- `UserProfile` - Profile information

### New Models

```go
// UserPrivacy - Privacy settings
type UserPrivacy struct {
    UserID   int64  `gorm:"primaryKey"`
    IsPublic bool   `gorm:"default:true"`            // Public/private account
    Settings string `gorm:"type:jsonb;default:'{}'"` // Reserved for future granular controls
}

// UserNotification - Notification settings
type UserNotification struct {
    UserID       int64  `gorm:"primaryKey"`
    PushEnabled  bool   `gorm:"default:true"`            // App push notifications
    EmailEnabled bool   `gorm:"default:true"`            // Email notifications
    Preferences  string `gorm:"type:jsonb;default:'{}'"` // Reserved for future notification types
}

// UserAddress - Shipping addresses
type UserAddress struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;index"`
    Name       string    `gorm:"type:varchar(50);not null"`  // Recipient name
    Phone      string    `gorm:"type:varchar(20);not null"`  // Phone number
    Province   string    `gorm:"type:varchar(50)"`           // Province
    City       string    `gorm:"type:varchar(50)"`           // City
    District   string    `gorm:"type:varchar(50)"`           // District
    Address    string    `gorm:"type:varchar(255);not null"` // Detailed address
    PostalCode string    `gorm:"type:varchar(20)"`           // Postal code
    IsDefault  bool      `gorm:"default:false"`              // Default address
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// AccountDeletion - Account deletion requests
type AccountDeletion struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    UserID      int64     `gorm:"not null;uniqueIndex"`
    Reason      string    `gorm:"type:varchar(255)"`  // Deletion reason
    ScheduledAt time.Time `gorm:"not null"`           // Scheduled deletion time (cooling period)
    CreatedAt   time.Time
}
```

## API Request/Response Structures

### Profile

```go
// GET /user/profile - Response
type ProfileResponse struct {
    UserID    int64  `json:"user_id"`
    Nickname  string `json:"nickname"`
    AvatarURL string `json:"avatar_url"`
    Intro     string `json:"intro"`
    Location  string `json:"location"`
}

// PATCH /user/profile - Request
type UpdateProfileRequest struct {
    Nickname  *string `json:"nickname,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
    Intro     *string `json:"intro,omitempty"`
    Location  *string `json:"location,omitempty"`
}
```

### Security

```go
// GET /user/security - Response
type SecurityOverviewResponse struct {
    HasPassword    bool     `json:"has_password"`
    LinkedAccounts []string `json:"linked_accounts"` // ["email", "google"]
    Email          string   `json:"email"`           // Masked: w***@gmail.com
}

// POST /user/security/password - Request
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}
```

### Privacy & Notifications

```go
// GET /user/privacy - Response
type PrivacyResponse struct {
    IsPublic bool `json:"is_public"`
}

// PATCH /user/privacy - Request
type UpdatePrivacyRequest struct {
    IsPublic *bool `json:"is_public,omitempty"`
}

// GET /user/notifications - Response
type NotificationResponse struct {
    PushEnabled  bool `json:"push_enabled"`
    EmailEnabled bool `json:"email_enabled"`
}

// PATCH /user/notifications - Request
type UpdateNotificationRequest struct {
    PushEnabled  *bool `json:"push_enabled,omitempty"`
    EmailEnabled *bool `json:"email_enabled,omitempty"`
}
```

### Addresses

```go
// GET /user/addresses - Response
type AddressListResponse struct {
    Addresses []AddressItem `json:"addresses"`
}

type AddressItem struct {
    ID         int64  `json:"id"`
    Name       string `json:"name"`
    Phone      string `json:"phone"`
    Province   string `json:"province"`
    City       string `json:"city"`
    District   string `json:"district"`
    Address    string `json:"address"`
    PostalCode string `json:"postal_code"`
    IsDefault  bool   `json:"is_default"`
}

// POST /user/addresses - Request
type CreateAddressRequest struct {
    Name       string `json:"name" binding:"required"`
    Phone      string `json:"phone" binding:"required"`
    Province   string `json:"province"`
    City       string `json:"city"`
    District   string `json:"district"`
    Address    string `json:"address" binding:"required"`
    PostalCode string `json:"postal_code"`
    IsDefault  bool   `json:"is_default"`
}
```

### Account Management

```go
// POST /user/account/export - Response
type ExportResponse struct {
    Message string `json:"message"`
}

// DELETE /user/account - Request
type DeleteAccountRequest struct {
    Password string `json:"password,omitempty"`
    Reason   string `json:"reason,omitempty"`
}

// DELETE /user/account - Response
type DeleteAccountResponse struct {
    Message     string    `json:"message"`
    ScheduledAt time.Time `json:"scheduled_at"`
}
```

## Business Logic & Edge Cases

### Security - Edge Cases

| Scenario | Handling |
|----------|----------|
| Change password - no existing password | OAuth user setting password first time, `old_password` not required |
| Change password - has existing password | Must verify `old_password` |
| Unlink OAuth - last login method | Reject: "Must keep at least one login method" |
| Unlink OAuth - only email left but no password | Reject: "Please set a password first" |
| Link OAuth - already linked to another account | Reject: "This account is already linked to another user" |

### Addresses - Edge Cases

| Scenario | Handling |
|----------|----------|
| Add address - first address | Automatically set as default |
| Delete address - deleting default | Automatically set oldest remaining as default |
| Set as default | Unset default flag on other addresses |
| Address limit | Limit to 20 addresses, return error if exceeded |

### Account Deletion - Flow

```
User requests deletion
    │
    ▼
Verify identity (password/OAuth)
    │
    ▼
Create AccountDeletion record
Set scheduled_at = now + 7 days
    │
    ▼
Send confirmation email
    │
    ▼
Set user status to "pending_deletion"
(Can still login, but show warning)
    │
    ▼
Can cancel within 7 days
    │
    ▼
Scheduled job performs actual deletion
```

### Data Export - Flow

```
User requests export
    │
    ▼
Create async job
    │
    ▼
Collect user data (profile, addresses, etc.)
    │
    ▼
Generate JSON/ZIP file
    │
    ▼
Upload to temporary storage (7-day expiry)
    │
    ▼
Send download link to user's email
```

## Implementation Notes

### File Structure

```
apps/core/internal/
├── handler/
│   ├── auth.go          # Existing auth handlers
│   ├── user.go          # New user handlers
│   └── routes.go        # Add /user routes
├── service/
│   ├── auth.go          # Existing auth service
│   └── user.go          # New user service
├── model/
│   └── user.go          # Add new models
└── dto/
    └── user.go          # Request/response structs
```

### Database Migrations

New tables to create:
- `user_privacies`
- `user_notifications`
- `user_addresses`
- `account_deletions`
