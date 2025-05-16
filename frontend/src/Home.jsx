
import React, { useEffect, useRef, useState } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import UserSidebar from './Sidebar/Sidebar';

const Home = () => {
  const { user, isAuthenticated, getAccessTokenSilently } = useAuth0();
  const [apiResponse, setApiResponse] = useState('');

  // Use a ref to track if we've already called the backend after login
  const calledBackendAfterLogin = useRef(false);

  useEffect(() => {
    const callBackend = async () => {
      try {
        const token = await getAccessTokenSilently();
        console.log('Access Token:', token); 

        const response = await fetch('http://localhost:8081/create', {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          setApiResponse('Failed to fetch data from backend');
          return;
        }

        const data = await response.json();
        setApiResponse(data.message || 'Backend data received');
      } catch (error) {
        setApiResponse('Error fetching data');
        console.error(error);
      }
    };

    // Only call backend once after login happens
    if (isAuthenticated && !calledBackendAfterLogin.current) {
      calledBackendAfterLogin.current = true;
      callBackend();
    }
  }, [isAuthenticated, getAccessTokenSilently]);

  return (
    <div style={{ display: 'flex' }}>
      <UserSidebar />
      <div style={{ textAlign: 'center', flex: 1, padding: '40px' }}>
        <h2>Welcome to CloudCord</h2>
        <p>Hello, {user?.nickname}!</p>
        <p>This is your home page after login.</p>
        {apiResponse && <p>Backend response: {apiResponse}</p>}
      </div>
    </div>
  );
};

export default Home;
