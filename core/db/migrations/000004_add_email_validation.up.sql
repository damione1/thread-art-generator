CREATE TABLE account_activations (
  id UUID DEFAULT uuid_generate_v1mc() PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_email VARCHAR(255) NOT NULL,
  activation_token INT NOT NULL,
  expiration TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '1 day',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX account_activations_user_id_idx ON account_activations(user_id);
CREATE INDEX account_activations_user_email_idx ON account_activations(user_email);
