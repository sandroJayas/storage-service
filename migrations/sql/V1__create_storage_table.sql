CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Storage Locations table
CREATE TABLE storage_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    capacity INT NOT NULL,
    current_load INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Boxes table
CREATE TABLE boxes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    packing_mode VARCHAR(20) NOT NULL CHECK (packing_mode IN ('self', 'sort')),
    status VARCHAR(30) NOT NULL CHECK (status IN ('in_transit', 'pending_pack', 'pending_pickup', 'stored', 'returned', 'disposed')),
    location_id UUID REFERENCES storage_locations(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Items table
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    box_id UUID NOT NULL REFERENCES boxes(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    quantity INT DEFAULT 1,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Storage Orders table
CREATE TABLE storage_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    box_id UUID NOT NULL REFERENCES boxes(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('pickup', 'return', 'relocate')),
    scheduled_date TIMESTAMP NOT NULL,
    status VARCHAR(30) NOT NULL CHECK (status IN ('requested', 'in_progress', 'completed', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
