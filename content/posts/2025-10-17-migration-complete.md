---
title: "Migration to File-Based Storage Complete"
date: 2025-10-17T14:30:00Z
slug: "migration-complete"
tags: ["technical", "updates"]
description: "Successfully migrated from PostgreSQL to file-based markdown storage"
published: true
---

# Migration Complete

The blog has been successfully migrated from a database-driven architecture to a static file-based system.

## Benefits

- **No database required** - Simplified deployment and maintenance
- **Version controlled content** - All posts are now in Git
- **Fast performance** - Posts loaded into memory on startup
- **Easy editing** - Just edit markdown files directly

## Technical Details

Posts are now stored as markdown files with YAML front matter in the `/content/posts/` directory. The system:

1. Loads all posts on startup
2. Renders markdown to HTML automatically
3. Sorts posts by date
4. Filters draft posts in production mode

Check back for more updates!
