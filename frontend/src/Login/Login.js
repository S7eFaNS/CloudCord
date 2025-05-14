import React from 'react';
import './Login.css';
import LoginButton from '../LoginButton';

const Login = () => {
  return (
    <div className="login">
      <h2>Welcome to CloudCord</h2>
      <div className="login__logo">
        <img src="https://logodownload.org/wp-content/uploads/2017/11/discord-logo-4.png" alt="CloudCord Logo" />
      </div>
      <LoginButton />
    </div>
  );
};

export default Login;