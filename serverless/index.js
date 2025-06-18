exports.friendAdded = (req, res) => {
  res.set('Access-Control-Allow-Origin', '*');
  res.set('Access-Control-Allow-Methods', 'POST, OPTIONS');
  res.set('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    return res.status(204).send('');
  }

  const { user_id, friend_id } = req.body;

  if (!user_id || !friend_id) {
    return res.status(400).json({ error: 'Missing user_id or friend_id' });
  }

  return res.status(200).json({
    notification: `User ${user_id} added ${friend_id} as a friend.`,
    friend_id,
  });
};

