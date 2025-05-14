
import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import UserSidebar from './Sidebar/Sidebar';

const Home = () => {
  const { user } = useAuth0();

  return (
    <div style={{ display: 'flex' }}>
      <UserSidebar />
      <div style={{ textAlign: 'center', flex: 1, padding: '40px' }}>
        <h2>Welcome to CloudCord</h2>
        <p>Hello, {user.nickname}!</p>
        <p>This is your home page after login.</p>
      </div>
    </div>
  );
};

export default Home;
