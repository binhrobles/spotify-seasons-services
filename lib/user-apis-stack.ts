import * as cdk from '@aws-cdk/core';
import * as golang from 'aws-lambda-golang';
import { LambdaRestApi, Cors } from '@aws-cdk/aws-apigateway';
import { Effect, PolicyStatement } from '@aws-cdk/aws-iam';

export class UserApisStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // user create/update function with env var input
    const userCrudFunction = new golang.GolangFunction(this, 'user-crud', {
      environment: {
        CLIENT_ID: process.env.CLIENT_ID || '',
        REDIRECT_URI: process.env.REDIRECT_URI || '',
        STAGE: 'production',
      },
    });

    // add ssm get-parameter permission and permission to decrypt
    userCrudFunction.addToRolePolicy(
      new PolicyStatement({
        effect: Effect.ALLOW,
        actions: ['ssm:GetParameter'],
        resources: [
          `arn:aws:ssm:${this.region}:${this.account}:parameter/spotifySeasons/clientSecret`,
        ],
      }),
    );
    userCrudFunction.addToRolePolicy(
      new PolicyStatement({
        effect: Effect.ALLOW,
        actions: ['kms:Decrypt'],
        resources: [`arn:aws:kms:${this.region}:${this.account}:alias/aws/ssm`],
      }),
    );

    // slap an API gateway in front of it
    const api = new LambdaRestApi(this, 'api', {
      handler: userCrudFunction,
      proxy: false,
      defaultCorsPreflightOptions: {
        allowOrigins: Cors.ALL_ORIGINS, // TODO: not this
        allowMethods: ['OPTIONS', 'PUT'],
      },
    });

    const user = api.root.addResource('user');
    user.addMethod('PUT');
  }
}
