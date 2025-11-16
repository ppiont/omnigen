import {
  CognitoUserPool,
  CognitoUser,
  AuthenticationDetails,
  CognitoUserAttribute,
} from "amazon-cognito-identity-js";

const poolData = {
  UserPoolId:
    import.meta.env.VITE_COGNITO_USER_POOL_ID || "us-east-1_PLACEHOLDER",
  ClientId: import.meta.env.VITE_COGNITO_CLIENT_ID || "PLACEHOLDER_CLIENT_ID",
};

const userPool = new CognitoUserPool(poolData);

/**
 * Login user with email and password
 * @param {string} email - User email
 * @param {string} password - User password
 * @returns {Promise<{accessToken: string, idToken: string, refreshToken: string}>}
 */
export function loginUser(email, password) {
  return new Promise((resolve, reject) => {
    const authenticationDetails = new AuthenticationDetails({
      Username: email,
      Password: password,
    });

    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.authenticateUser(authenticationDetails, {
      onSuccess: (result) => {
        const tokens = {
          accessToken: result.getAccessToken().getJwtToken(),
          idToken: result.getIdToken().getJwtToken(),
          refreshToken: result.getRefreshToken().getToken(),
        };
        resolve(tokens);
      },
      onFailure: (err) => {
        reject(err);
      },
      newPasswordRequired: (userAttributes) => {
        // User needs to set a new password
        reject({
          code: "NewPasswordRequired",
          message: "Please set a new password",
          userAttributes,
        });
      },
    });
  });
}

/**
 * Sign up new user
 * @param {string} name - User's full name
 * @param {string} email - User email
 * @param {string} password - User password
 * @returns {Promise<{userSub: string, userConfirmed: boolean}>}
 */
export function signupUser(name, email, password) {
  return new Promise((resolve, reject) => {
    const attributeList = [
      new CognitoUserAttribute({
        Name: "email",
        Value: email,
      }),
      new CognitoUserAttribute({
        Name: "name",
        Value: name,
      }),
    ];

    userPool.signUp(email, password, attributeList, null, (err, result) => {
      if (err) {
        reject(err);
        return;
      }

      resolve({
        userSub: result.userSub,
        userConfirmed: result.userConfirmed,
      });
    });
  });
}

/**
 * Confirm user signup with verification code
 * @param {string} email - User email
 * @param {string} code - Verification code from email
 * @returns {Promise<string>} Success message
 */
export function confirmSignup(email, code) {
  return new Promise((resolve, reject) => {
    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.confirmRegistration(code, true, (err, result) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(result);
    });
  });
}

/**
 * Resend verification code
 * @param {string} email - User email
 * @returns {Promise<string>}
 */
export function resendConfirmationCode(email) {
  return new Promise((resolve, reject) => {
    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.resendConfirmationCode((err, result) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(result);
    });
  });
}

/**
 * Initiate forgot password flow
 * @param {string} email - User email
 * @returns {Promise<any>}
 */
export function forgotPassword(email) {
  return new Promise((resolve, reject) => {
    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.forgotPassword({
      onSuccess: (data) => {
        resolve(data);
      },
      onFailure: (err) => {
        reject(err);
      },
    });
  });
}

/**
 * Confirm new password with verification code
 * @param {string} email - User email
 * @param {string} code - Verification code from email
 * @param {string} newPassword - New password
 * @returns {Promise<string>}
 */
export function confirmPassword(email, code, newPassword) {
  return new Promise((resolve, reject) => {
    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.confirmPassword(code, newPassword, {
      onSuccess: () => {
        resolve("Password reset successful");
      },
      onFailure: (err) => {
        reject(err);
      },
    });
  });
}

/**
 * Get current authenticated user session
 * @returns {Promise<{accessToken: string, idToken: string, refreshToken: string}>}
 */
export function getCurrentSession() {
  return new Promise((resolve, reject) => {
    const cognitoUser = userPool.getCurrentUser();

    if (!cognitoUser) {
      reject(new Error("No user found"));
      return;
    }

    cognitoUser.getSession((err, session) => {
      if (err) {
        reject(err);
        return;
      }

      if (!session.isValid()) {
        reject(new Error("Session is invalid"));
        return;
      }

      resolve({
        accessToken: session.getAccessToken().getJwtToken(),
        idToken: session.getIdToken().getJwtToken(),
        refreshToken: session.getRefreshToken().getToken(),
      });
    });
  });
}

/**
 * Refresh user session using refresh token
 * @returns {Promise<{accessToken: string, idToken: string}>}
 */
export function refreshSession() {
  return new Promise((resolve, reject) => {
    const cognitoUser = userPool.getCurrentUser();

    if (!cognitoUser) {
      reject(new Error("No user found"));
      return;
    }

    cognitoUser.getSession((err, session) => {
      if (err) {
        reject(err);
        return;
      }

      const refreshToken = session.getRefreshToken();

      cognitoUser.refreshSession(refreshToken, (refreshErr, newSession) => {
        if (refreshErr) {
          reject(refreshErr);
          return;
        }

        resolve({
          accessToken: newSession.getAccessToken().getJwtToken(),
          idToken: newSession.getIdToken().getJwtToken(),
          refreshToken: newSession.getRefreshToken().getToken(),
        });
      });
    });
  });
}

/**
 * Sign out current user
 */
export function signOut() {
  const cognitoUser = userPool.getCurrentUser();
  if (cognitoUser) {
    cognitoUser.signOut();
  }
}

/**
 * Get user attributes
 * @returns {Promise<Object>} User attributes
 */
export function getUserAttributes() {
  return new Promise((resolve, reject) => {
    const cognitoUser = userPool.getCurrentUser();

    if (!cognitoUser) {
      reject(new Error("No user found"));
      return;
    }

    cognitoUser.getSession((err) => {
      if (err) {
        reject(err);
        return;
      }

      cognitoUser.getUserAttributes((attrErr, attributes) => {
        if (attrErr) {
          reject(attrErr);
          return;
        }

        const attributesObj = {};
        attributes.forEach((attr) => {
          attributesObj[attr.Name] = attr.Value;
        });

        resolve(attributesObj);
      });
    });
  });
}
