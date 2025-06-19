import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 500,
  duration: '120s'
};

export default function () {
  let res = http.get('https://cloudcord.info/message/chat?user1=auth0|6824b28215aa6ecd9d4f9305&user2=auth0|682d05e0575b4c2f4bce');

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}

