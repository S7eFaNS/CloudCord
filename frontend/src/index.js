import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { Auth0Provider } from '@auth0/auth0-react';
import { BrowserRouter } from 'react-router';

const domain = "dev-p3oldabcwb4l1kia.us.auth0.com";
const clientId = "yvYQ5suAYzulGEpcCVFavrpugQOYzbmW";

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
<Auth0Provider
    domain={domain}
    clientId={clientId}
    authorizationParams={{
      redirect_uri: window.location.origin
    }}
    cacheLocation="localstorage"
    useRefreshTokens={true}
  >
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </Auth0Provider>
);

