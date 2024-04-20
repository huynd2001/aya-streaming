import { defineEventHandler, getQuery, getRequestHeader } from 'h3';

import axios from 'axios';
import { constants } from '../../app/interfaces/constants';

const endPoint = import.meta.env[constants.API_URL_ENV];

export default defineEventHandler(async (event) => {
  const params = getQuery(event);
  const authorization = getRequestHeader(event, 'Authorization');

  try {
    let { data } = await axios.delete(`${endPoint}/session/`, {
      headers: {
        Authorization: authorization,
      },
      params: params,
    });
    return data;
  } catch (err) {
    if (axios.isAxiosError(err)) {
      if (err.response) {
        console.error(err.response.status);
        return err.response.data;
      }
    } else {
      console.error(err);
      return { err: err };
    }
  }
});
