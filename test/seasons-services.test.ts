import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as SeasonsServices from '../lib/seasons-services-stack';

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new SeasonsServices.SeasonsServicesStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});
