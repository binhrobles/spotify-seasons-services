import * as cdk from '@aws-cdk/core';
import * as golang from 'aws-lambda-golang';
import { LambdaRestApi } from '@aws-cdk/aws-apigateway';

export class SeasonsServicesStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const userCrudFunction = new golang.GolangFunction(this, 'user-crud', {});

    const api = new LambdaRestApi(this, 'api', {
      handler: userCrudFunction,
      proxy: false,
    });

    const user = api.root.addResource('user');
    user.addMethod('POST');
  }
}
