-- Accounts: users, their sessions, and a log of sign-in attempts.

CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- Stored trimmed and in lowercase, so You@Mail.com and you@mail.com are
    -- one account. The CHECK refuses anything the application forgot to
    -- lowercase, and UNIQUE refuses a second account for the same address.
    email         TEXT NOT NULL UNIQUE CHECK (email = lower(email)),
    -- A bcrypt hash, never the password itself. The hash includes its own
    -- random salt and its cost setting.
    password_hash TEXT NOT NULL,
    -- The demo account's password is printed in the README. It is marked so
    -- it can be left out of the per-email sign-in limit and found by the seed.
    is_demo       BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A session is one signed-in browser. The browser holds a random token in a
-- cookie. This table holds only the SHA-256 hash of that token, so someone
-- who reads the table cannot use what they find to sign in.
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Used when a user signs in, to delete their expired sessions.
CREATE INDEX sessions_user_idx ON sessions (user_id);

-- One row per event that is rate limited. Serverless functions share no
-- memory, so the counting has to happen in the database.
--   signin_email: a failed sign-in, key = the email that was tried
--   signin_ip:    a failed sign-in, key = the address it came from
--   signup_ip:    a sign-up,        key = the address it came from
CREATE TABLE auth_attempts (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind         TEXT NOT NULL CHECK (kind IN ('signin_email', 'signin_ip', 'signup_ip')),
    key          TEXT NOT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The limiter asks "how many rows of this kind and key since time T".
CREATE INDEX auth_attempts_lookup_idx ON auth_attempts (kind, key, attempted_at);
