
import { useAuth0 } from '@auth0/auth0-react';
import './Sidebar.css';
import { Link } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import LogoutButton from '../LogoutButton'; // adjust the path if necessary
import DeleteButton from '../DeleteButton';

  //const apiDevUrl = 'http://localhost:8081';
  const apiProdUrl = 'https://cloudcord.com/user';
  

const UserSidebar = () => {
  const { user, getAccessTokenSilently, isAuthenticated } = useAuth0();
  const [users, setUsers] = useState([]);
  const [currentUserId, setCurrentUserId] = useState(null);
  const [friendStatuses, setFriendStatuses] = useState({});
  
  
  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchCurrentUser = async () => {
      try {
        const token = await getAccessTokenSilently();
        const res = await fetch(`${apiProdUrl}/auth-user?auth0_id=${encodeURIComponent(user.sub)}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (!res.ok) throw new Error('Failed to get current user');
        const data = await res.json();
        setCurrentUserId(data.userID);
      } catch (error) {
        console.error(error);
      }
    };

    fetchCurrentUser();
  }, [user, getAccessTokenSilently, isAuthenticated]);

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
  
  useEffect(() => {
    if (!currentUserId || users.length === 0) return;

    const checkFriendshipStatuses = async () => {
      try {
        const token = await getAccessTokenSilently();

        const statuses = {};

        await Promise.all(users.map(async (friend) => {
          if (friend.user_id === currentUserId) return; 
          const res = await fetch(`${apiProdUrl}/is-friend?user_id=${currentUserId}&other_id=${friend.user_id}`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          if (!res.ok) {
            console.error(`Failed to check friend status for user ${friend.user_id}`);
            statuses[friend.user_id] = false;
            return;
          }
          const data = await res.json();
          statuses[friend.user_id] = data.are_friends;
        }));
        
        setFriendStatuses(statuses);
      } catch (error) {
        console.error('Error checking friend statuses:', error);
      }
    };

    checkFriendshipStatuses();
  }, [currentUserId, users, getAccessTokenSilently]);

  const handleAddFriend = async (friendUserId) => {
    if (!currentUserId) {
      console.error('No current user ID');
      return;
    }

    try {
      const token = await getAccessTokenSilently();

      const response = await fetch(`${apiProdUrl}/add-friend`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: currentUserId,
          friend_id: friendUserId,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to add friend');
      }

      alert('Friend added successfully!');

     setFriendStatuses(prev => ({ ...prev, [friendUserId]: true }));
    } catch (error) {
      console.error('Add friend error:', error);
    }
  };

  return (
    <div className="userSidebar">
      <div className="userSidebar__top">
        <div className="userSidebar__topRow">
          <h3>{user?.nickname}</h3>
          <LogoutButton />
          <DeleteButton />
        </div>
      </div>

      <div className="userSidebar__list">
        {users.length === 0 ? (
          <p>Loading users...</p>
        ) : (
          users
            .filter(u => u.user_id !== currentUserId)
            .map(u => (
              <div
                key={u.user_id}
                className="userSidebar__userRow"
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              >
                <Link
                  to={`/chat?user1=${encodeURIComponent(user.sub)}&user2=${encodeURIComponent(u.auth0_id)}`}
                  className="userSidebar__user"
                >
                  <p>{u.username}</p>
                </Link>

                {friendStatuses[u.user_id] === false ? (
                  <button onClick={() => handleAddFriend(u.user_id)}>Add Friend</button>
                ) : friendStatuses[u.user_id] === true ? (
                  <span style={{ color: 'green', fontWeight: 'bold' }}>Friend</span>
                ) : null}
              </div>
            ))
        )}
      </div>
    </div>
  );
};

export default UserSidebar;