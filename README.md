# s3-copy
Tool for faster transfer multiple files to AWS S3, by using GO routines.

## Install

```bash
go get github.com/sarunask/s3-copy
```

## Usage

Dry run, with debug, nothing uploaded. You could check if exclude works as expected.
```bash
./s3-copy --path ~/Music/ --s3-bucket=ssss --dry-run --debug --sse-c-key 45123qwefawdfgddddadfqwefgqwegdd --exclude '.*\.mp4' --workers 50
```