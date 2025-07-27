# V-Cloud Backend Database Queries Documentation

## Database Information
- **Database Name**: `vcloud`
- **Connection**: `mysql://root:vithu@localhost:3306/vcloud`
- **Tables**: `user`, `metadata`

## All Database Queries Used in the Application

### 1. User Authentication Queries

#### Get User for Login (getUser function)
```sql
SELECT user_id, email, password FROM user WHERE email = ?
```
**Purpose**: Retrieve user credentials for authentication
**Parameters**: email (string)

#### Register User (registerUser function)
```sql
CALL addUser(?,?);
```
**Purpose**: Register a new user using stored procedure
**Parameters**: email (string), hashed_password (string)

#### Get User ID by Email (getuseridformail function)
```sql
SELECT user_id FROM user WHERE email = ?
```
**Purpose**: Find user ID by email address
**Parameters**: email (string)

#### Get User Storage (multiple functions)
```sql
SELECT used_storage FROM user WHERE user_id = ?
```
**Purpose**: Get current storage usage for a user
**Parameters**: user_id (string)

### 2. File Metadata Queries

#### Check File Exists (addfileentry function)
```sql
SELECT COUNT(*) FROM metadata WHERE name_at_server = ? AND owner = ?
```
**Purpose**: Check if file already exists for a user
**Parameters**: name_at_server (string), owner (string)

#### Insert File Entry (addfileentry function)
```sql
INSERT INTO metadata (name_at_server, actualname, filetype, file_size, is_accepted, owner) 
VALUES (?,?,?,?,?,?);
```
**Purpose**: Add new file metadata to database
**Parameters**: name_at_server, actualname, filetype, file_size, is_accepted, owner

#### Get File Location (getfileloc function)
```sql
SELECT name_at_server, filetype FROM metadata WHERE id = ?
```
**Purpose**: Get file server name and type for download
**Parameters**: id (int)

#### Get User Files (getfiles function)
```sql
SELECT id, actualname, filetype, file_size, date FROM metadata 
WHERE owner = ? AND is_accepted = ?;
```
**Purpose**: Get all files for a user (accepted or pending)
**Parameters**: owner_id (string), accepted (boolean)

#### Get Single File (getfile function)
```sql
SELECT id, name_at_server, actualname, filetype, file_size, date FROM metadata 
WHERE id = ?;
```
**Purpose**: Get complete file information by ID
**Parameters**: file_id (string)

### 3. File Management Queries

#### Accept File (acceptfile function)
```sql
UPDATE metadata SET is_accepted = true WHERE id = ?
```
**Purpose**: Accept a shared file
**Parameters**: id (string)

#### Reject File (rejectfile function)
```sql
DELETE FROM metadata WHERE id = ?
```
**Purpose**: Reject/delete a shared file
**Parameters**: id (string)

#### Delete File (deletefile function)
```sql
DELETE FROM metadata WHERE id = ?
```
**Purpose**: Delete file metadata from database
**Parameters**: id (string)

#### Check File Reference Count (deletefile function)
```sql
SELECT COUNT(*) FROM metadata WHERE name_at_server = ?
```
**Purpose**: Check if other users have the same file before physical deletion
**Parameters**: name_at_server (string)

## Database Schema Requirements

### User Table Structure
```sql
CREATE TABLE user (
    user_id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    used_storage DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Metadata Table Structure
```sql
CREATE TABLE metadata (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name_at_server VARCHAR(255) NOT NULL,
    actualname VARCHAR(255) NOT NULL,
    filetype VARCHAR(50) NOT NULL,
    file_size DECIMAL(10,2) NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    owner VARCHAR(255) NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner) REFERENCES user(user_id) ON DELETE CASCADE
);
```

### Required Stored Procedure
```sql
DELIMITER //
CREATE PROCEDURE addUser(
    IN p_email VARCHAR(255),
    IN p_password VARCHAR(255)
)
-- Implementation in database_setup.sql
```

## API Endpoints and Their Database Operations

| Endpoint | Method | Database Operations |
|----------|--------|-------------------|
| `/signup` | POST | `CALL addUser(?,?)` |
| `/signin` | POST | `SELECT user_id, email, password FROM user WHERE email = ?` |
| `/get-files` | POST | `SELECT id, actualname, filetype, file_size, date FROM metadata WHERE owner = ? AND is_accepted = ?` |
| `/get-shared-files` | POST | `SELECT id, actualname, filetype, file_size, date FROM metadata WHERE owner = ? AND is_accepted = ?` |
| `/download` | POST | `SELECT name_at_server, filetype FROM metadata WHERE id = ?` |
| `/complete-upload` | POST | `INSERT INTO metadata...`, `SELECT used_storage FROM user WHERE user_id = ?` |
| `/share` | POST | `SELECT id, name_at_server... FROM metadata WHERE id = ?`, `SELECT user_id FROM user WHERE email = ?`, `INSERT INTO metadata...` |
| `/accept` | POST | `UPDATE metadata SET is_accepted = true WHERE id = ?` |
| `/reject` | POST | `DELETE FROM metadata WHERE id = ?` |
| `/deletefile` | POST | `DELETE FROM metadata WHERE id = ?`, `SELECT COUNT(*) FROM metadata WHERE name_at_server = ?` |

## Performance Considerations

### Recommended Indexes
```sql
CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_metadata_owner_accepted ON metadata(owner, is_accepted);
CREATE INDEX idx_name_server_owner ON metadata(name_at_server, owner);
```

### Storage Tracking
The application tracks user storage in the `used_storage` field. Consider implementing triggers to automatically update this field when files are added, deleted, or accepted.

## Security Notes

1. **Password Hashing**: The application uses bcrypt for password hashing
2. **JWT Tokens**: Uses JWT for session management with 72-hour expiry
3. **File Access Control**: Files are linked to users through the `owner` field
4. **SQL Injection Prevention**: All queries use parameterized statements

## File System Integration

- **Upload Directory**: `./uploads/{uuid}/` (temporary storage during upload)
- **Final Storage**: `./userfiles/{uuid}` (permanent file storage)
- **File Sharing**: Multiple metadata entries can reference the same physical file
- **File Deletion**: Physical files are only deleted when no metadata entries reference them
