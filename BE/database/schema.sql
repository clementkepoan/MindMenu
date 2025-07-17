-- Create restaurants table
CREATE TABLE IF NOT EXISTS restaurants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Add reference to auth.users if using Supabase Auth
    CONSTRAINT fk_owner FOREIGN KEY (owner_id) REFERENCES auth.users(id) ON DELETE CASCADE
);

-- Create branches table
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    restaurant_id UUID NOT NULL,
    name TEXT NOT NULL,
    address TEXT,
    has_chatbot BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
);

-- Create chatbots table
CREATE TABLE IF NOT EXISTS chatbots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'building', -- 'active', 'building', 'error'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

-- Create a view to get all chatbots with restaurant and branch info
CREATE OR REPLACE VIEW chatbot_info AS
SELECT 
    c.id as chatbot_id,
    c.status,
    c.created_at as chatbot_created,
    b.id as branch_id,
    b.name as branch_name,
    r.id as restaurant_id,
    r.name as restaurant_name,
    r.owner_id
FROM chatbots c
JOIN branches b ON c.branch_id = b.id
JOIN restaurants r ON b.restaurant_id = r.id;

-- Create RLS (Row Level Security) policies for tables
-- This ensures users can only access their own data

-- Restaurants: users can only select, update, and delete their own restaurants
ALTER TABLE restaurants ENABLE ROW LEVEL SECURITY;

CREATE POLICY restaurants_select_policy ON restaurants 
    FOR SELECT USING (owner_id = auth.uid());

CREATE POLICY restaurants_insert_policy ON restaurants 
    FOR INSERT WITH CHECK (owner_id = auth.uid());

CREATE POLICY restaurants_update_policy ON restaurants 
    FOR UPDATE USING (owner_id = auth.uid());

CREATE POLICY restaurants_delete_policy ON restaurants 
    FOR DELETE USING (owner_id = auth.uid());

-- Branches: users can only manage branches of their restaurants
ALTER TABLE branches ENABLE ROW LEVEL SECURITY;

CREATE POLICY branches_select_policy ON branches 
    FOR SELECT USING (
        restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = auth.uid())
    );

CREATE POLICY branches_insert_policy ON branches 
    FOR INSERT WITH CHECK (
        restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = auth.uid())
    );

CREATE POLICY branches_update_policy ON branches 
    FOR UPDATE USING (
        restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = auth.uid())
    );

CREATE POLICY branches_delete_policy ON branches 
    FOR DELETE USING (
        restaurant_id IN (SELECT id FROM restaurants WHERE owner_id = auth.uid())
    );

-- Chatbots: users can only manage chatbots for their branches
ALTER TABLE chatbots ENABLE ROW LEVEL SECURITY;

CREATE POLICY chatbots_select_policy ON chatbots 
    FOR SELECT USING (
        branch_id IN (
            SELECT b.id FROM branches b
            JOIN restaurants r ON b.restaurant_id = r.id
            WHERE r.owner_id = auth.uid()
        )
    );

CREATE POLICY chatbots_insert_policy ON chatbots 
    FOR INSERT WITH CHECK (
        branch_id IN (
            SELECT b.id FROM branches b
            JOIN restaurants r ON b.restaurant_id = r.id
            WHERE r.owner_id = auth.uid()
        )
    );

CREATE POLICY chatbots_update_policy ON chatbots 
    FOR UPDATE USING (
        branch_id IN (
            SELECT b.id FROM branches b
            JOIN restaurants r ON b.restaurant_id = r.id
            WHERE r.owner_id = auth.uid()
        )
    );

CREATE POLICY chatbots_delete_policy ON chatbots 
    FOR DELETE USING (
        branch_id IN (
            SELECT b.id FROM branches b
            JOIN restaurants r ON b.restaurant_id = r.id
            WHERE r.owner_id = auth.uid()
        )
    );

-- Create indexes for better query performance
CREATE INDEX idx_branches_restaurant_id ON branches(restaurant_id);
CREATE INDEX idx_chatbots_branch_id ON chatbots(branch_id);
