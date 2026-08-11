\set ON_ERROR_STOP on

BEGIN;

\ir schema/001_users.sql
\ir schema/002_photos.sql
\ir schema/003_conversations.sql
\ir schema/004_chat_messages.sql
\ir schema/005_notifications.sql
\ir schema/006_profile_decisions.sql
\ir schema/007_profile_views.sql
\ir schema/008_user_blocks.sql
\ir schema/009_user_reports.sql
\ir schema/010_account_tokens.sql
\ir schema/011_auth_sessions.sql

COMMIT;