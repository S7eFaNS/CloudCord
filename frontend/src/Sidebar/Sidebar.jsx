
import { useAuth0 } from '@auth0/auth0-react';
import './Sidebar.css';
import { Link } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import LogoutButton from '../LogoutButton'; // adjust the path if necessary
import DeleteButton from '../DeleteButton';

  //const apiDevUrl = 'http://localhost:8081';
  const apiProdUrl = 'https://cloudcord.info/user';
  

const UserSidebar = () => {
  const { user, getAccessTokenSilently, isAuthenticated } = useAuth0();
  const [users, setUsers] = useState([]);
  const [currentUserId, setCurrentUserId] = useState(null);
  const [friendStatuses, setFriendStatuses] = useState({});
  const [recommendations, setRecommendations] = useState([]);
  const [view, setView] = useState('all');
  
  
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
  

const fetchRecommendations = async () => {
  try {
    if (!currentUserId) return;
    const token = await getAccessTokenSilently();

    const response = await fetch(`${apiProdUrl}/recommendations?user_id=${currentUserId}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch recommendations');
    }

    const data = await response.json();
    setRecommendations(data);
  } catch (error) {
    console.error('Error fetching recommendations:', error);
  }
};


const handleToggleView = (viewName) => {
  setView(viewName);
  if (viewName === 'recommendations' && recommendations.length === 0) {
    fetchRecommendations();
  }
};

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

    const friendAddedNotification = await fetch(
      'https://us-central1-directed-sonar-461707-r8.cloudfunctions.net/friendAdded',
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: currentUserId,
          friend_id: friendUserId,
        }),
      }
    );

    const notifyData = await friendAddedNotification.json();
    alert(notifyData.notification);

      setFriendStatuses(prev => ({ ...prev, [friendUserId]: true }));
      setRecommendations(prev => prev.filter(rec => rec.ID !== friendUserId));
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




<div className="toggleButtons">
  <button
    className={`toggleButton ${view === 'all' ? 'active' : ''}`}
    onClick={() => handleToggleView('all')}
  >
    All Users
  </button>
  <button
    className={`toggleButton ${view === 'recommendations' ? 'active' : ''}`}
    onClick={() => handleToggleView('recommendations')}
  >
    Recommendations
  </button>
</div>

      {view === 'all' && (
        <div className="userSidebar__list">
          <h4>All Users</h4>
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
      )}


      {view === 'recommendations' && currentUserId && (
        <div className="userSidebar__recommendations">
          <h4>Friend Recommendations</h4>
          {recommendations.length === 0 ? (
            <p>No recommendations available</p>
          ) : (
            recommendations.map(rec => (
              <div
                key={rec.id}
                className="userSidebar__userRow"
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              >
                <p>{rec.username}</p>
                {friendStatuses[rec.id] ? (
                  <span style={{ color: 'green', fontWeight: 'bold' }}>Friend</span>
                ) : (
                  <button onClick={() => handleAddFriend(rec.id)}>Add Friend</button>
                )}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};


export default UserSidebar;