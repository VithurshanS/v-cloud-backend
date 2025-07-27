-- V-Cloud Backend Database Setup Script
-- Database: vcloud
-- This script creates all the necessary tables and stored procedures for the V-Cloud application

-- Create the database
CREATE DATABASE IF NOT EXISTS vcloud;
USE vcloud;

-- Drop tables if they exist (for clean setup)
DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS user;

-- Create user table
CREATE TABLE user (
    user_id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    used_storage DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create metadata table for file information
CREATE TABLE metadata (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name_at_server VARCHAR(255) NOT NULL,
    actualname VARCHAR(255) NOT NULL,
    filetype VARCHAR(50) NOT NULL,
    file_size DECIMAL(10,2) NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    owner VARCHAR(255) NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (owner) REFERENCES user(user_id) ON DELETE CASCADE,
    INDEX idx_owner (owner),
    INDEX idx_name_server_owner (name_at_server, owner),
    INDEX idx_is_accepted (is_accepted)
);

-- Create indexes for better performance
CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_metadata_owner_accepted ON metadata(owner, is_accepted);

-- Stored procedure to add a new user
DELIMITER //
CREATE PROCEDURE addUser(
    IN p_email VARCHAR(255),
    IN p_password VARCHAR(255)
)
BEGIN
    DECLARE user_count INT DEFAULT 0;
    DECLARE new_user_id VARCHAR(255);
    
    -- Check if email already exists
    SELECT COUNT(*) INTO user_count FROM user WHERE email = p_email;
    
    IF user_count > 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Email already exists';
    ELSE
        -- Generate a unique user ID (you might want to use UUID() function)
        SET new_user_id = CONCAT('user_', UNIX_TIMESTAMP(), '_', CONNECTION_ID());
        
        -- Insert new user
        INSERT INTO user (user_id, email, password, used_storage) 
        VALUES (new_user_id, p_email, p_password, 0.00);
        
        -- Return the new user ID
        SELECT new_user_id as user_id, p_email as email;
    END IF;
END //
DELIMITER ;

-- Trigger to update used_storage when files are added
DELIMITER //
CREATE TRIGGER update_storage_on_insert 
    AFTER INSERT ON metadata 
    FOR EACH ROW 
BEGIN
    IF NEW.is_accepted = TRUE THEN
        UPDATE user 
        SET used_storage = used_storage + NEW.file_size 
        WHERE user_id = NEW.owner;
    END IF;
END //
DELIMITER ;

-- Trigger to update used_storage when files are deleted
DELIMITER //
CREATE TRIGGER update_storage_on_delete 
    AFTER DELETE ON metadata 
    FOR EACH ROW 
BEGIN
    IF OLD.is_accepted = TRUE THEN
        UPDATE user 
        SET used_storage = used_storage - OLD.file_size 
        WHERE user_id = OLD.owner;
    END IF;
END //
DELIMITER ;

-- Trigger to update used_storage when file acceptance status changes
DELIMITER //
CREATE TRIGGER update_storage_on_accept 
    AFTER UPDATE ON metadata 
    FOR EACH ROW 
BEGIN
    -- If file was just accepted
    IF OLD.is_accepted = FALSE AND NEW.is_accepted = TRUE THEN
        UPDATE user 
        SET used_storage = used_storage + NEW.file_size 
        WHERE user_id = NEW.owner;
    END IF;
    
    -- If file was rejected (from accepted to not accepted)
    IF OLD.is_accepted = TRUE AND NEW.is_accepted = FALSE THEN
        UPDATE user 
        SET used_storage = used_storage - NEW.file_size 
        WHERE user_id = NEW.owner;
    END IF;
END //
DELIMITER ;

-- Insert sample data (optional - remove if not needed)
-- INSERT INTO user (user_id, email, password, used_storage) 
-- VALUES ('test_user_1', 'test@example.com', '$2a$10$example_hashed_password', 0.00);

-- View to get user storage summary
CREATE VIEW user_storage_summary AS
SELECT 
    u.user_id,
    u.email,
    u.used_storage,
    COUNT(m.id) as total_files,
    COUNT(CASE WHEN m.is_accepted = TRUE THEN 1 END) as accepted_files,
    COUNT(CASE WHEN m.is_accepted = FALSE THEN 1 END) as pending_files
FROM user u
LEFT JOIN metadata m ON u.user_id = m.owner
GROUP BY u.user_id, u.email, u.used_storage;

-- Show tables and their structure
SHOW TABLES;
DESCRIBE user;
DESCRIBE metadata;
