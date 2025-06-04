import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';

const apiProdUrl = "https://cloudcord.com/user";

const DeleteButton = () => {
  const { user, getAccessTokenSilently, logout } = useAuth0();

  const handleDelete = async () => {
    if (!window.confirm("Are you sure you want to delete your account? This action is irreversible.")) {
      return;
    }

    try {
      const token = await getAccessTokenSilently();

      const response = await fetch(`${apiProdUrl}/delete?auth0_id=${encodeURIComponent(user.sub)}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to delete user");
      }

      alert("Your account has been deleted.");
      logout({ logoutParams: { returnTo: window.location.origin } });

    } catch (error) {
      console.error("Error deleting user:", error);
      alert("There was a problem deleting your account.");
    }
  };

  return (
    <button onClick={handleDelete} style={{ color: "red" }}>
      Delete My Account
    </button>
  );
};

export default DeleteButton;