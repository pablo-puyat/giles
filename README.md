# Giles Digital Asset Manager (DAM) CLI

This Digital Asset Manager CLI provides a set of tools for managing, organizing, and retrieving digital files. Built using the Cobra CLI framework in Go.

## Features

- File scanning and hashing
- File organization
- File search
- File retrieval

## Installation

TBD

## Usage

The Giles DAM CLI offers the following main commands:

### `scan`

Scan files in a specified directory, computing hashes and storing metadata.

```
giles scan <source path> 
```

Examples:
```
giles scan /path/to/source
```

### `organize`

Move and organize files into a new directory structure based on their hash.

```
giles organize <source> <destination>
```

Example:
```
giles organize /path/to/source /path/to/destination
```
