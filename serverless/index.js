exports.friendAdded = (req, res) => {
  res.set('Access-Control-Allow-Origin', '*');
  res.set('Access-Control-Allow-Methods', 'POST, OPTIONS');
  res.set('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    return res.status(204).send('');
  }

  const { friend_username } = req.body;

  if (!friend_username) {
    return res.status(400).json({ error: 'Missing friend_username' });
  }

  return res.status(200).json({
    notification: `Added ${friend_username} as a friend.`,
  });
};

