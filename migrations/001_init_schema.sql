-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- 'admin' or 'user'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Create attendance_locations table
CREATE TABLE IF NOT EXISTS attendance_locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    radius INTEGER DEFAULT 10, -- in meters
    is_active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create spatial index for better performance
CREATE INDEX IF NOT EXISTS idx_attendance_locations_coords ON attendance_locations
USING GIST (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326));

-- Create attendances table
CREATE TABLE IF NOT EXISTS attendances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location_id INTEGER NOT NULL REFERENCES attendance_locations(id) ON DELETE RESTRICT,
    check_in_time TIMESTAMP NOT NULL,
    check_out_time TIMESTAMP,
    check_in_latitude DECIMAL(10, 8) NOT NULL,
    check_in_longitude DECIMAL(11, 8) NOT NULL,
    check_out_latitude DECIMAL(10, 8),
    check_out_longitude DECIMAL(11, 8),
    distance_from_location DECIMAL(10, 2), -- in meters
    status VARCHAR(20) DEFAULT 'present', -- 'present', 'late', 'half_day'
    notes TEXT,
    photo_url VARCHAR(500), -- optional selfie photo
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for attendances
CREATE INDEX IF NOT EXISTS idx_attendances_user_date ON attendances(user_id, DATE(check_in_time));
CREATE INDEX IF NOT EXISTS idx_attendances_location ON attendances(location_id);
CREATE INDEX IF NOT EXISTS idx_attendances_check_in_time ON attendances(check_in_time);

-- Create work_schedules table
CREATE TABLE IF NOT EXISTS work_schedules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    check_in_start TIME NOT NULL, -- e.g., 08:00:00
    check_in_end TIME NOT NULL,   -- e.g., 09:00:00 (late after this)
    check_out_start TIME NOT NULL, -- e.g., 17:00:00
    work_days INTEGER[], -- [1,2,3,4,5] for Mon-Fri (1=Monday, 7=Sunday)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_schedules table
CREATE TABLE IF NOT EXISTS user_schedules (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT,
    location_id INTEGER NOT NULL REFERENCES attendance_locations(id) ON DELETE RESTRICT,
    effective_from DATE NOT NULL,
    effective_to DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, effective_from)
);

-- Create indexes for user_schedules
CREATE INDEX IF NOT EXISTS idx_user_schedules_user ON user_schedules(user_id);
CREATE INDEX IF NOT EXISTS idx_user_schedules_dates ON user_schedules(effective_from, effective_to);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_attendance_locations_updated_at BEFORE UPDATE ON attendance_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_attendances_updated_at BEFORE UPDATE ON attendances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_work_schedules_updated_at BEFORE UPDATE ON work_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: admin123)
-- Password hash is bcrypt hash of 'admin123'
INSERT INTO users (email, password_hash, full_name, role)
VALUES (
    'admin@attendance.com',
    '$2a$10$8K1p/a0dL3LKzIKjvL5zvOqR0K9Gw7RBUOLVQHfNLHPDJLGKLqZC2',
    'System Administrator',
    'admin'
) ON CONFLICT (email) DO NOTHING;

-- Insert sample work schedule
INSERT INTO work_schedules (name, check_in_start, check_in_end, check_out_start, work_days)
VALUES (
    'Standard Office Hours',
    '08:00:00',
    '09:00:00',
    '17:00:00',
    ARRAY[1,2,3,4,5]
) ON CONFLICT DO NOTHING;
