import {
  expect as expectCDK,
  matchTemplate,
  MatchStyle,
} from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as UserApis from '../lib/user-apis-stack';

test('Empty Stack', () => {
  const app = new cdk.App();
  // WHEN
  const stack = new UserApis.UserApisStack(app, 'MyTestStack');
  // THEN
  expectCDK(stack).to(
    matchTemplate(
      {
        Resources: {},
      },
      MatchStyle.EXACT,
    ),
  );
});
