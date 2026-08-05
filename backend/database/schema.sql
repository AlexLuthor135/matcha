\set ON_ERROR_STOP on

BEGIN;

\ir schema/001_users.sql
\ir schema/002_photos.sql
\ir schema/003_conversations.sql
\ir schema/004_chat_messages.sql
\ir schema/005_notifications.sql
\ir schema/006_profile_decisions.sql

COMMIT;