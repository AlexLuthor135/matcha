import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import axiosInstance from "../AxiosInstance";
import { useAuth } from "../AuthProvider";
import { useWebSocket } from "../WebSocketProvider";
import { Button } from "../components/Button";
import "./Chat.css";

function getPeerID(conversation, currentUserID) {
    if (!conversation || !currentUserID) {
        return null;
    }
    return conversation.user_one_id === currentUserID
        ? conversation.user_two_id
        : conversation.user_one_id;
}

function formatMessageTime(value) {
    if (!value) {
        return "";
    }
    return new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
        minute: "2-digit",
    }).format(new Date(value));
}

export default function Chat() {
    const { userID } = useAuth();
    const { send, subscribe, status } = useWebSocket();
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedUserParam = searchParams.get("user") || "";
    const [conversations, setConversations] = useState([]);
    const [activeConversationID, setActiveConversationID] = useState(null);
    const [recipientID, setRecipientID] = useState(selectedUserParam);
    const [messages, setMessages] = useState([]);
    const [draft, setDraft] = useState("");
    const [error, setError] = useState("");
    const messagesEndRef = useRef(null);
    const pendingReadMessageIDsRef = useRef(new Set());

    const activeConversation = useMemo(
        () => conversations.find((conversation) => conversation.id === activeConversationID) || null,
        [activeConversationID, conversations]
    );

    const activeRecipientID = useMemo(() => {
        if (recipientID) {
            return Number(recipientID);
        }
        return getPeerID(activeConversation, userID);
    }, [activeConversation, recipientID, userID]);

    useEffect(() => {
        setRecipientID(selectedUserParam);
    }, [selectedUserParam]);

    useEffect(() => {
        const loadConversations = async () => {
            try {
                const response = await axiosInstance.get("/backend/api/accounts/conversations");
                const items = Array.isArray(response.data?.conversations)
                    ? response.data.conversations
                    : [];

                setConversations(items);

                if (!selectedUserParam && items.length > 0) {
                    setActiveConversationID(items[0].id);
                }
            } catch (requestError) {
                console.log("CONVERSATIONS ERROR", requestError);
                setError("Could not load conversations");
            }
        };

        loadConversations();
    }, [selectedUserParam]);

    useEffect(() => {
        if (!activeConversationID) {
            setMessages([]);
            return;
        }

        const loadMessages = async () => {
            try {
                const response = await axiosInstance.get(
                    `/backend/api/accounts/conversations/${activeConversationID}/messages`
                );
                setMessages(Array.isArray(response.data?.messages) ? response.data.messages : []);
            } catch (requestError) {
                console.log("MESSAGES ERROR", requestError);
                setError("Could not load messages");
            }
        };

        loadMessages();
    }, [activeConversationID]);

    useEffect(() => {
        const unsubscribeChat = subscribe("chat_message", (payload) => {
            const nextMessage = {
                id: payload.id,
                conversation_id: payload.conversation_id,
                sender_id: payload.sender_id,
                recipient_id: payload.recipient_id,
                content: payload.message,
                created_at: payload.created_at,
                updated_at: payload.created_at,
                read_at: null,
            };

            setMessages((currentMessages) => {
                const belongsToOpenConversation =
                    payload.conversation_id === activeConversationID;
                const isOwnMessage = payload.sender_id === userID;

                if (!belongsToOpenConversation && !isOwnMessage) {
                    return currentMessages;
                }
                if (currentMessages.some((message) => message.id === nextMessage.id)) {
                    return currentMessages;
                }
                return [...currentMessages, nextMessage];
            });

            setConversations((currentConversations) => {
                const userOneID = Math.min(payload.sender_id, payload.recipient_id);
                const userTwoID = Math.max(payload.sender_id, payload.recipient_id);
                const nextConversation = {
                    id: payload.conversation_id,
                    user_one_id: userOneID,
                    user_two_id: userTwoID,
                    updated_at: payload.created_at,
                };
                const existingConversations = currentConversations.filter(
                    (conversation) => conversation.id !== payload.conversation_id
                );
                return [nextConversation, ...existingConversations];
            });

            if (payload.sender_id === userID) {
                setActiveConversationID(payload.conversation_id);
                setRecipientID(String(payload.recipient_id));
                setSearchParams({ user: String(payload.recipient_id) });
            }
        });

        const unsubscribeMessageRead = subscribe("message_read", (payload) => {
            pendingReadMessageIDsRef.current.delete(payload.message_id);

            setMessages((currentMessages) => currentMessages.map((message) => (
                message.id === payload.message_id
                    ? { ...message, read_at: payload.read_at }
                    : message
            )));
        });

        const unsubscribeError = subscribe("error", (payload) => {
            setError(payload.message || "Chat error");
        });

        return () => {
            unsubscribeChat();
            unsubscribeMessageRead();
            unsubscribeError();
        };
    }, [activeConversationID, setSearchParams, subscribe, userID]);

    useEffect(() => {
        if (status !== "connected" || !activeConversationID || !userID) {
            if (status !== "connected") {
                pendingReadMessageIDsRef.current.clear();
            }
            return;
        }

        messages.forEach((message) => {
            const isUnreadIncomingMessage =
                message.conversation_id === activeConversationID &&
                message.recipient_id === userID &&
                !message.read_at;

            if (
                !isUnreadIncomingMessage ||
                pendingReadMessageIDsRef.current.has(message.id)
            ) {
                return;
            }

            const isSent = send({
                type: "message_read",
                message_id: message.id,
            });

            if (isSent) {
                pendingReadMessageIDsRef.current.add(message.id);
            }
        });
    }, [activeConversationID, messages, send, status, userID]);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const selectConversation = (conversation) => {
        const peerID = getPeerID(conversation, userID);
        setActiveConversationID(conversation.id);
        setRecipientID(peerID ? String(peerID) : "");
        setSearchParams(peerID ? { user: String(peerID) } : {});
    };

    const sendMessage = (event) => {
        event.preventDefault();
        const content = draft.trim();
        const parsedRecipientID = Number(activeRecipientID);

        if (!content || !parsedRecipientID) {
            return;
        }

        const isSent = send({
            type: "chat_message",
            recipient_id: parsedRecipientID,
            message: content,
        });

        if (!isSent) {
            setError("Chat is not connected yet");
            return;
        }

        setDraft("");
        setError("");
    };

    return (
        <div className="chat-page">
            <aside className="chat-sidebar" aria-label="Conversations">
                <div className="chat-sidebar-header">
                    <p className="chat-kicker">{status}</p>
                    <h2>Chat</h2>
                </div>

                <label className="chat-recipient-label" htmlFor="chat-recipient">
                    Recipient ID
                </label>
                <input
                    id="chat-recipient"
                    className="chat-recipient-input"
                    value={recipientID}
                    inputMode="numeric"
                    placeholder="User ID"
                    onChange={(event) => {
                        setRecipientID(event.target.value.replace(/\D/g, ""));
                        setActiveConversationID(null);
                    }}
                />

                <div className="chat-conversations">
                    {conversations.length > 0 ? conversations.map((conversation) => {
                        const peerID = getPeerID(conversation, userID);
                        const isActive = conversation.id === activeConversationID;
                        return (
                            <button
                                key={conversation.id}
                                className={`chat-conversation${isActive ? " is-active" : ""}`}
                                type="button"
                                onClick={() => selectConversation(conversation)}
                            >
                                <span>User #{peerID}</span>
                                <small>Open conversation</small>
                            </button>
                        );
                    }) : (
                        <p className="chat-empty">No conversations yet</p>
                    )}
                </div>
            </aside>

            <section className="chat-panel" aria-label="Messages">
                <div className="chat-panel-header">
                    <div>
                        <p className="chat-kicker">Conversation</p>
                        <h2>{activeRecipientID ? `User #${activeRecipientID}` : "Select a user"}</h2>
                    </div>
                </div>

                <div className="chat-messages">
                    {messages.length > 0 ? messages.map((message) => {
                        const isMine = message.sender_id === userID;
                        return (
                            <div
                                key={message.id}
                                className={`chat-message${isMine ? " is-mine" : ""}`}
                            >
                                <p>{message.content}</p>
                                <div className="chat-message-meta">
                                    <time>{formatMessageTime(message.created_at)}</time>
                                    {isMine ? (
                                        <span>{message.read_at ? "Read" : "Sent"}</span>
                                    ) : null}
                                </div>
                            </div>
                        );
                    }) : (
                        <p className="chat-empty chat-empty-centered">Start the conversation</p>
                    )}
                    <div ref={messagesEndRef} />
                </div>

                {error ? <p className="chat-error">{error}</p> : null}

                <form className="chat-form" onSubmit={sendMessage}>
                    <textarea
                        value={draft}
                        placeholder="Write a message..."
                        rows="2"
                        onChange={(event) => setDraft(event.target.value)}
                        onKeyDown={(event) => {
                            if (event.key === "Enter" && !event.shiftKey) {
                                sendMessage(event);
                            }
                        }}
                    />
                    <Button type="submit" disabled={!draft.trim() || !activeRecipientID}>
                        Send
                    </Button>
                </form>
            </section>
        </div>
    );
}
