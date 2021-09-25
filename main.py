from boto.s3.connection import S3Connection
import subprocess
import os


s3 = S3Connection(os.environ['DISCORD_TOKEN'])
print(f"Debug: Discord key: {os.environ['DISCORD_TOKEN']}")
subprocess.run(f"go run discordtyping.go -t {os.environ['DISCORD_TOKEN']}")
