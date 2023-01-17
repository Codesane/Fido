# Fido
A very simple watchdog written in Go.

### Motivation
When working with servies outside of a managed cluster such as Kubernetes or AWS ECS where the orchestrator itself manages restarts and availability I found the need for a very simple application that restarts and notifies our Slack channel whenever the application crashes. Fido has been used in production for several years successfully.

### Usage
Run the build.sh script to produce the binary, modify the flags in the script if you want to compile for a differnet archutecture.

Run the resulting binary like so:
`./fido <command to be watched>` e.g. `./fido npm run start`

You also have to configure a slack webhook that Fido will report to. This can be done by setting the environment variable: `FIDO_SLACK_WEB_HOOK` to the full url to the slack webhook.

### Cautions
There is no way to know if the watchdog itself crashes (I have not seen any reports of this happening as of yet). To reduce the time to action it would be a good idea to include some form of monitor for the system. For web servies this would be a pinging service that can notify operations if the service should become unreachable.
