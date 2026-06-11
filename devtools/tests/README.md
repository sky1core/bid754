# IBM decTest Inputs

This directory is populated by:

```bash
make setup-generation-inputs
```

The generated verification pipeline scans official `*.decTest` files from this
directory. The data files are intentionally not tracked; the setup script
downloads `dectest.zip`, verifies its checksum, and extracts it here.
