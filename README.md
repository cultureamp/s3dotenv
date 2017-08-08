s3dotenv
========

`s3dotenv` wraps a program with extra environment variables that it downloads from an S3 object specified by the `S3DOTENV` environment variable. This makes it a useful container `ENTRYPOINT` for environments like [Amazon ECS][ecs] where support for configuration is quite poor. If `S3DOTENV` isn't set, `s3dotenv` gets out of the way, just executing the program.

Create a env file in S3 (you'll need to create the bucket, permissions etc first):

```sh
echo "EXAMPLE_FOO=remote" | aws --region=us-west-2 s3 cp --sse=aws:kms - s3://your-bucket/path/to/file.env
```

Run a program (in this case `/usr/bin/env`) with the additional environment:

```sh
export EXAMPLE_BAR=local
export S3DOTENV="s3://your-bucket/path/to/file.env?region=us-west-2"
s3dotenv env | grep EXAMPLE

# 2017/08/02 17:43:12 loading env from s3://your-bucket/path/to/file.env?region=us-west-2
# EXAMPLE_BAR=local
# EXAMPLE_FOO=remote
```

Use it as a `Dockerfile` `ENTRYPOINT`:

```Dockerfile
COPY s3dotenv /usr/local/bin/s3dotenv
ENTRYPOINT ["/usr/local/bin/s3dotenv"]
```

Or as a `Dockerfile` `CMD` wrapper:

```Dockerfile
COPY s3dotenv /usr/local/bin/s3dotenv
CMD ["/usr/local/bin/s3dotenv", "your-existing-cmd", "and", "args"]
```

AWS credentials are discovered in the usual way by the [AWS SDK for Go][aws-sdk]. On AWS, instance/task IAM roles should be used. For local environments, consider [`aws-vault`][aws-vault].

Local environment variables take precedence over those in the env file; if an environment variable exists locally (even if it's blank) the value in the env file will be ignored.

The env file is parsed by [joho/godotenv][godotenv] (a Go port of [bkeepers/dotenv][dotenv]); here's an example:

```
# This is a comment
ACME_API_TOKEN=abc123
PASSWORD='pas$word'
PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\nHkVN9…\n-----END DSA PRIVATE KEY-----\n"
SECRET_HASH="something-with-a-#-hash"
SECRET_KEY=YOURSECRETKEYGOESHERE # inline comment
```

[aws-vault]: https://github.com/99designs/aws-vault
[godotenv]: https://github.com/joho/godotenv
[aws-sdk]: https://aws.amazon.com/sdk-for-go/
[dotenv]: https://github.com/bkeepers/dotenv
[ecs]: https://aws.amazon.com/ecs/
