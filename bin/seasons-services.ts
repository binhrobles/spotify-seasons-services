#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { SeasonsServicesStack } from '../lib/seasons-services-stack';

const app = new cdk.App();
new SeasonsServicesStack(app, 'SeasonsServicesStack');
