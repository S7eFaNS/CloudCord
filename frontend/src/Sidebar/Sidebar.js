
import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import './Sidebar.css';
import LogoutButton from '../LogoutButton'; // adjust the path if necessary

const dummyUsers = [
  { id: 1, name: 'Alice Johnson' },
  { id: 2, name: 'Bob Smith' },
  { id: 3, name: 'Charlie Rose' },
];

const UserSidebar = () => {
  const { user } = useAuth0();

  return (
    <div className="userSidebar">
      <div className="userSidebar__top">
        <div className="userSidebar__topRow">
          <h3>{user.nickname}</h3>
          <LogoutButton />
        </div>
      </div>

      <div className="userSidebar__list">
        {dummyUsers.map(user => (
          <div key={user.id} className="userSidebar__user">
            <p>{user.name}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default UserSidebar;

