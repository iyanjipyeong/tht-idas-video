INSERT INTO tiers (id, name, level, price, currency, description) VALUES
('11111111-1111-4111-8111-111111111101', 'Bronze', 1, 50000.00, 'IDR', 'Can access Bronze videos only'),
('11111111-1111-4111-8111-111111111102', 'Silver', 2, 100000.00, 'IDR', 'Can access Silver and Bronze videos'),
('11111111-1111-4111-8111-111111111103', 'Gold', 3, 150000.00, 'IDR', 'Can access Gold, Silver, and Bronze videos');

INSERT INTO users (name, email, password) VALUES ('Demo User', 'demo@example.com', '$2a$12$iBec91px7V2VgfAnSrmFkuzZJ.faY2nVw/L1ZEy1reZ3saB5YPqSW');

INSERT INTO videos (title, description, tier_id, video_url) VALUES
('Bronze Video', 'Video for Bronze tier', '11111111-1111-4111-8111-111111111101', 'https://example.com/bronze.mp4'),
('Silver Video', 'Video for Silver tier', '11111111-1111-4111-8111-111111111102', 'https://example.com/silver.mp4'),
('Gold Video', 'Video for Gold tier', '11111111-1111-4111-8111-111111111103', 'https://example.com/gold.mp4');
