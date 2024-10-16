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
giles scan [directory] [flags]
```

Flags:
- `--organize` or `-o`: Organize files after scanning
- `--destination` or `-d`: Destination directory for organized files

Examples:
```
giles scan /path/to/source
giles scan /path/to/source --organize --destination /path/to/dest
```

### `organize`

Move and organize files into a new directory structure based on their hash.

```
giles organize [source] [destination]
```

Example:
```
giles organize /path/to/source /path/to/destination
```

### `search`

Search for files based on various criteria.

```
giles search [flags]
```

Flags:
- `--type` or `-t`: File type to search for
- `--name` or `-n`: File name to search for

Example:
```
giles search --type pdf --name report
```

### `retrieve`

Retrieve file information or download files.

```
giles retrieve [file-id] [flags]
```

Flags:
- `--download` or `-d`: Download the file instead of returning its location

Example:
```
giles retrieve 123456 --download
```
