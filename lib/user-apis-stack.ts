import * as cdk from '@aws-cdk/core';
import * as golang from 'aws-lambda-golang';
import { LambdaRestApi, Cors } from '@aws-cdk/aws-apigateway';

export class UserApisStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const userCrudFunction = new golang.GolangFunction(this, 'user-crud', {
      environment: {
        CLIENT_ID: process.env.CLIENT_ID || '',
        REDIRECT_URI: process.env.REDIRECT_URI || '',
      },
    });

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
