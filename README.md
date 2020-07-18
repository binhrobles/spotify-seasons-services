# Spotify Seasons Services

## Stack

- Golang lambdas
- Create/Update User API accepts Spotify auth code and saves user to DynamoDB
- CDK for infrastructure declaration/deployment
- local development with SAM toolkit

## Useful commands

- `yarn build` compile typescript to js
- `yarn watch` watch for changes and compile
- `yarn test` perform the jest unit tests
- `cdk deploy` deploy this stack to your default AWS account/region
- `cdk diff` compare deployed stack with current state
- `cdk synth` emits the synthesized CloudFormation template
- `yarn local-invoke` synthesizes the current CFN template and calls `sam invoke` user-crud function
