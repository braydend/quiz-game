import { store } from "../store/appStore";
import type { Message } from "../types";

export let socket: WebSocket|undefined = undefined

export const createSocket = () => {
    socket = new WebSocket(`wss://${location.host}/ws/join`);

    return socket
}

export const createSocketFromId = (id:string) => {
    socket = new WebSocket(`wss://${location.host}/ws/join/${id}`);


    socket.onmessage = (e) => {
        const parsedMessage = JSON.parse(e.data) as Message;
        console.log(parsedMessage);
        if (parsedMessage.command.startsWith("SYS")) {
          handleSystemMessage(parsedMessage);
          return;
        }
  
      };
  
      socket.addEventListener("open", () => {
        handleSend({ command: "SYS_SYNC" });
      });

    return socket
}

export const handleSend = (msg: Message) => {
    socket?.send(JSON.stringify(msg));
  };

const handleSystemMessage = (msg: Message) => {
    const { command, payload } = msg;

    console.log("SYSTEM MESSAGE:", msg);

    switch (command) {
    case "SYS_READY":
        store.isReady.set(payload)
        break;
    case "SYS_UPDATE_NAME":
        store.name.set(payload);
        break;
    case "SYS_CORRECT_ANSWER":
        if (payload) {
            const parsedPayload = payload;
            const { name, sprite } = parsedPayload;
            store.correctAnswers.update((prev) => {
            prev[name] = sprite;
            return prev;
            });
        }
        break;
    case "SYS_CLEAR_PROMPT":
        store.prompt.set("");
        break;
    case "SYS_PROMPT_PROGRESS":
        store.promptProgress.set({
            total: payload.totalAnswers,
            remaining: payload.remainingAnswers,
        });
        break;
    case "SYS_PROMPT":
        store.prompt.set(payload);
        break;
    case "SYS_UPDATE_LEADERBOARD":
        store.leaderboard.set(payload.scores);
        break;
    case "SYS_UPDATE_USER_DATA":
        store.id.set(payload.id);
        store.name.set(payload.name);
        break;
    default:
        console.error(`Unknown system command: ${msg}`);
    }
};