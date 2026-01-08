# Rockbot Lyrics Service

A Go service that fetches, parses, and enriches song lyrics data.

## Status

🚧 Work in Progress

## Overview

This service provides structured, enriched lyrics data by:
- Fetching lyrics from LRCLib API
- Parsing timestamp data
- Detecting song structure (chorus)
- Calculating statistics
- Caching results for performance

## Tech Stack

- Go 1.21+
- Redis (caching with LFU eviction)
- Docker