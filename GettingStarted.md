# Getting Started with the SystemLink CLI

The SystemLink CLI allows you to call the SystemLink tag and message APIs from the command line. This makes it easy and convenient to automate API calls through scripts.

The CLI works natively on windows (systemlink.exe), linux (systemlink) and mac os (systemlink.osx). No install or any runtime required.

It is generated based on swagger YAML files and can be easily extended by just dropping new swagger files in the models directory. They are automatically picked up and displayed.

## Prerequisites 

Download the SystemLink CLI:

https://systemlink-releases.s3.amazonaws.com/systemlink-cli/systemlink-cli.zip

**Optional:**

The examples use "jq" which is a command-line JSON processor, you can install it:

Ubuntu:  `sudo apt install jq`  
MacOS:   `brew install jq`  
Windows: `chocolatey install jq` (or download it here https://stedolan.github.io/jq/download/)  

## How to configure the CLI

- Generate an API key
  - Go to: https://www.systemlinkcloud.com/security
  - Click "+ NEW API KEY"
  - Copy & Paste the generated API key
  - Make sure that the policies of your API key provide the necessary permissions

- Configure the CLI to use the new API key by default:

You can configure the global CLI options by creating the systemlink.yaml file in your home directory (or next to the executable). The following script sets up the default profile with an API key in the systemlink.yaml:

```bash
cat <<EOT > ~/systemlink.yaml
---
  profiles:
    - name: default
      api-key: <put your api key here>
EOT
```

Alternative: You can also execute the following command to set your API key as an environment variable for your current terminal session:
```bash
export NI_API_KEY=<put your api key here>
```

- Make sure that the cli works. The following command should return all your tags (or at least [] if you don't have any tags yet)
```bash
./systemlink tags get-tags
```

## How to send messages?

```bash
token=$(./systemlink messages create-session | jq -r '.token')
./systemlink messages subscribe-to-topic --token $token --topic mytopic
./systemlink messages publish-message --token $token --topic mytopic --message hello
./systemlink messages read-message --token $token --timeoutMilliseconds 10000 
```

## How to create tags?

```bash
./systemlink tags create-tag --path "mytag" --properties "{}" --type "DOUBLE" --collectAggregates true --keywords "foo,bar"
./systemlink tags get-tags
./systemlink tags get-tag --path "mytag"
```

## Need help about the supported operations?

List all message service operations:
```bash
./systemlink messages --help
```

List all tag service operations:
```bash
./systemlink tags --help
```

List all parameters for a specific operations:
```bash
./systemlink tags create-selection --help
```

## Which parameters are supported?

Simple types (strings, integers, floating point numbers and booleans):

```bash
--path "mytag"
--collectAggregates true
--skip 5
--average 7.51
```

Array types (string, integer, floating point and boolean arrays). Simply separate the array elements with a comma:
```bash
--keywords "foo,bar"
```

Complex types, using JSON:
```bash
--properties '{ "my-values": ["a", "b", "c"] }'
```

## How to set up a profile in the configuration file?

Create a new "systemlink.yaml" file in the home directory or next to the executable. The yaml file supports the following
parameters:

```yaml
---
profiles:
  - name: default # default profile, used when no --profile parameter specified
    api-key: <your default API key>       # The x-ni-api-key of your default profile
  - name: my-profile
    api-key: <my api key>                 # The x-ni-api-key
    url: https://api.systemlinkcloud.com  # Base url for all HTTP requests
    insecure: true                        # Ignores SSL certificate errors
    verbose: true                         # Outputs full request and response, used for debugging
```

You can use the profile with the name "default" to specifiy parameters which should be included when you omit the --profile flag.

```bash
./systemlink tags get-tags --profile my-profile
```
