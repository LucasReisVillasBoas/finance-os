INSERT INTO categories (id, user_id, name, type, icon, color, is_system) VALUES
-- Despesas
(uuid_generate_v4(), NULL, 'Alimentação', 'expense', 'restaurant', '#FF5722', TRUE),
(uuid_generate_v4(), NULL, 'Transporte', 'expense', 'directions_car', '#2196F3', TRUE),
(uuid_generate_v4(), NULL, 'Moradia', 'expense', 'home', '#9C27B0', TRUE),
(uuid_generate_v4(), NULL, 'Saúde', 'expense', 'local_hospital', '#F44336', TRUE),
(uuid_generate_v4(), NULL, 'Educação', 'expense', 'school', '#3F51B5', TRUE),
(uuid_generate_v4(), NULL, 'Lazer', 'expense', 'sports_esports', '#00BCD4', TRUE),
(uuid_generate_v4(), NULL, 'Roupas', 'expense', 'checkroom', '#E91E63', TRUE),
(uuid_generate_v4(), NULL, 'Tecnologia', 'expense', 'devices', '#607D8B', TRUE),
(uuid_generate_v4(), NULL, 'Assinaturas', 'expense', 'subscriptions', '#FF9800', TRUE),
(uuid_generate_v4(), NULL, 'Outros', 'expense', 'more_horiz', '#9E9E9E', TRUE),
-- Receitas
(uuid_generate_v4(), NULL, 'Salário', 'income', 'work', '#4CAF50', TRUE),
(uuid_generate_v4(), NULL, 'Freelance', 'income', 'laptop', '#8BC34A', TRUE),
(uuid_generate_v4(), NULL, 'Investimentos', 'income', 'trending_up', '#CDDC39', TRUE),
(uuid_generate_v4(), NULL, 'Aluguel Recebido', 'income', 'house', '#FFEB3B', TRUE),
(uuid_generate_v4(), NULL, 'Outros', 'income', 'attach_money', '#9E9E9E', TRUE);
