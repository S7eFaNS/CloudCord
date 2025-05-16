import React from 'react';
import { Route, Routes, Navigate } from 'react-router';
import { useAuth0 } from '@auth0/auth0-react';
import Home from './Home';
import Login from './Login/Login';
import ChatPage from './Chat/Chatpage'

function App() {
  const { isAuthenticated, isLoading } = useAuth0();

  if (isLoading) return <div style={{ textAlign: 'center', marginTop: '50px' }}>Loading...</div>;

  return (
    <div>
      {isAuthenticated}

      <Routes>
        <Route
          path="/"
          element={isAuthenticated ? <Navigate to="/home" /> : <Login />}
        />
        <Route
          path="/home"
          element={isAuthenticated ? <Home /> : <Navigate to="/" />}
        />
        <Route
          path="/chat"
          element={isAuthenticated ? <ChatPage receiverName="User123" /> : <Navigate to="/" />}
        />
      </Routes>
    </div>
  );
}

export default App;