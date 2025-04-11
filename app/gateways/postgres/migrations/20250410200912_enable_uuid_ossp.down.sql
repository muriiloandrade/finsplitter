-- Down Migration: Remove the uuid-ossp extension if it exists.

DROP EXTENSION IF EXISTS "uuid-ossp";