import type { Message } from "../types";

export let socket: WebSocket|undefined = undefined

export const createSocket = () => {
    socket = new WebSocket(`wss://${location.host}/ws/join`);

    return socket
}

export const handleSend = (msg: Message) => {
    socket?.send(JSON.stringify(msg));
  };