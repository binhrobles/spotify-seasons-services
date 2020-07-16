#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { UserApisStack } from '../lib/user-apis-stack';

const app = new cdk.App();
new UserApisStack(app, 'UserApisStack');
