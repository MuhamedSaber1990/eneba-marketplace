-- Profiles: a name, when the terms were accepted, and the marketing choice.
-- Only the users table changes. No session is touched, so everyone who is
-- signed in stays signed in.

-- The name shown in the header and on the account page.
--
-- The DEFAULT exists for one reason: this migration runs a few minutes
-- before the new code is live. In that gap the old code still creates
-- accounts without a name, and they must not fail. The new code always
-- sends a name, so the default is never used after that.
ALTER TABLE users ADD COLUMN name TEXT NOT NULL DEFAULT 'Player';

-- Accounts that exist already never typed a name. They get the part of
-- their email before the @, cut to the longest name allowed. The demo
-- account gets a proper one.
UPDATE users
   SET name = CASE
                WHEN is_demo THEN 'Demo Player'
                ELSE COALESCE(NULLIF(left(split_part(email, '@', 1), 60), ''), 'Player')
              END;

-- The same limit the application checks. char_length counts characters, not
-- bytes, so a name in any alphabet gets the same room.
ALTER TABLE users ADD CONSTRAINT users_name_length
    CHECK (char_length(name) BETWEEN 1 AND 60);

-- When this account ticked "I agree to the Terms and Privacy Policy".
-- NULL for every account made before this migration: they were never shown
-- the checkbox, and the database does not claim that they agreed.
ALTER TABLE users ADD COLUMN terms_accepted_at TIMESTAMPTZ;

-- "Send me news about deals". Off unless the person switched it on. Nothing
-- is ever sent: the project has no email service. It is a stored preference.
ALTER TABLE users ADD COLUMN marketing_opt_in BOOLEAN NOT NULL DEFAULT false;
