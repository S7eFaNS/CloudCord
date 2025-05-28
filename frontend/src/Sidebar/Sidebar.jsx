
import { useAuth0 } from '@auth0/auth0-react';
import './Sidebar.css';
import { Link } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import LogoutButton from '../LogoutButton'; // adjust the path if necessary

  //const apiDevUrl = 'http://localhost:8081';
  const apiProdUrl = 'https://cloudcord.com/user';
  

const UserSidebar = () => {
  const { user, getAccessTokenSilently, isAuthenticated } = useAuth0();
  const [users, setUsers] = useState([]);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const token = await getAccessTokenSilently();

        const response = await fetch(`${apiProdUrl}/users`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) throw new Error('Failed to fetch users');

        const data = await response.json();
        setUsers(data);
      } catch (error) {
        console.error('Error fetching users:', error);
      }
    };

    if (isAuthenticated) {
      fetchUsers();
    }
  }, [getAccessTokenSilently, isAuthenticated]);

  return (
    <div className="userSidebar">
      <div className="userSidebar__top">
        <div className="userSidebar__topRow">
          <h3>{user?.nickname}</h3>
          <LogoutButton />
        </div>
      </div>

      <div className="userSidebar__list">
        {users.length === 0 ? (
          <p>Loading users...</p>
        ) : (
        users
          .filter(u => u.auth0_id !== user.sub)
          .map(u => (
            <Link
              key={u.user_id}
              to={`/chat?user1=${encodeURIComponent(user.sub)}&user2=${encodeURIComponent(u.auth0_id)}`}
              className="userSidebar__user"
            >
              <p>{u.username}</p>
            </Link>
        ))
      )}
      </div>
    </div>
  );
};

export default UserSidebar;