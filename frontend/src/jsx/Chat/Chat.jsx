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

function normalizeMatch(match) {
    if (!match || typeof match !== "object") {
        return null;
    }

    const id = Number(match.id);
    if (!Number.isInteger(id) || id < 1) {
        return null;
    }

    return {
        id,
        userName: String(match.user_name ?? ""),
        firstName: String(match.first_name ?? ""),
        lastName: String(match.last_name ?? ""),
        avatar: String(match.avatar ?? ""),
    };
}

function getMatchName(match) {
    if (!match) {
        return "";
    }

    return [match.firstName, match.lastName].filter(Boolean).join(" ")
        || match.userName
        || `User #${match.id}`;
}

function getAvatarURL(avatar) {
    return avatar ? `/backend${avatar}` : "/1.png";
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
    const [matches, setMatches] = useState([]);
    const [conversations, setConversations] = useState([]);
    const [activeConversationID, setActiveConversationID] = useState(null);
    const [recipientID, setRecipientID] = useState(null);
    const [messages, setMessages] = useState([]);
    const [draft, setDraft] = useState("");
    const [error, setError] = useState("");
    const [isLoadingSidebar, setIsLoadingSidebar] = useState(true);
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

    const matchesByUserID = useMemo(
        () => new Map(matches.map((match) => [match.id, match])),
        [matches]
    );

    const activeMatch = matchesByUserID.get(Number(activeRecipientID)) || null;
    const activeRecipientName = activeMatch
        ? getMatchName(activeMatch)
        : activeRecipientID
            ? `User #${activeRecipientID}`
            : "Select a match";

    useEffect(() => {
        if (!userID) {
            return undefined;
        }

        let isCancelled = false;

        const loadSidebar = async () => {
            try {
                const [matchesResponse, conversationsResponse] = await Promise.all([
                    axiosInstance.get("/backend/api/accounts/matches"),
                    axiosInstance.get("/backend/api/accounts/conversations"),
                ]);

                if (isCancelled) {
                    return;
                }

                const matchItems = Array.isArray(matchesResponse.data?.matches)
                    ? matchesResponse.data.matches.map(normalizeMatch).filter(Boolean)
                    : [];
                const conversationItems = Array.isArray(conversationsResponse.data?.conversations)
                    ? conversationsResponse.data.conversations
                    : [];

                setMatches(matchItems);
                setConversations(conversationItems);
            } catch (requestError) {
                if (!isCancelled) {
                    console.log("CHAT SIDEBAR ERROR", requestError);
                    setError("Could not load matches and conversations");
                }
            } finally {
                if (!isCancelled) {
                    setIsLoadingSidebar(false);
                }
            }
        };

        loadSidebar();

        return () => {
            isCancelled = true;
        };
    }, [userID]);

    useEffect(() => {
        if (isLoadingSidebar || !userID) {
            return;
        }

        const selectedUserID = Number(selectedUserParam);
        if (Number.isInteger(selectedUserID) && selectedUserID > 0) {
            const selectedConversation = conversations.find(
                (conversation) => getPeerID(conversation, userID) === selectedUserID
            );
            const isKnownMatch = matches.some((match) => match.id === selectedUserID);

            if (selectedConversation || isKnownMatch) {
                setRecipientID(selectedUserID);
                setActiveConversationID(selectedConversation?.id ?? null);
                return;
            }

            setRecipientID(null);
            setActiveConversationID(null);
            setError("This user is not an active match");
            return;
        }

        if (conversations.length > 0) {
            const firstConversation = conversations[0];
            setActiveConversationID(firstConversation.id);
            setRecipientID(getPeerID(firstConversation, userID));
            return;
        }

        setActiveConversationID(null);
        setRecipientID(null);
    }, [conversations, isLoadingSidebar, matches, selectedUserParam, userID]);

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
                const messagePeerID = isOwnMessage
                    ? payload.recipient_id
                    : payload.sender_id;
                const belongsToSelectedMatch =
                    !activeConversationID && messagePeerID === activeRecipientID;

                if (!belongsToOpenConversation && !isOwnMessage && !belongsToSelectedMatch) {
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

            const messagePeerID = payload.sender_id === userID
                ? payload.recipient_id
                : payload.sender_id;

            if (payload.sender_id === userID || messagePeerID === activeRecipientID) {
                setActiveConversationID(payload.conversation_id);
                setRecipientID(messagePeerID);
                setSearchParams({ user: String(messagePeerID) });
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
    }, [activeConversationID, activeRecipientID, setSearchParams, subscribe, userID]);

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
        setRecipientID(peerID || null);
        setSearchParams(peerID ? { user: String(peerID) } : {});
        setError("");
    };

    const selectMatch = (match) => {
        const existingConversation = conversations.find(
            (conversation) => getPeerID(conversation, userID) === match.id
        );

        setRecipientID(match.id);
        setActiveConversationID(existingConversation?.id ?? null);
        setSearchParams({ user: String(match.id) });
        setError("");
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

                {isLoadingSidebar ? (
                    <p className="chat-empty">Loading chat...</p>
                ) : (
                    <>
                        <section className="chat-sidebar-section" aria-labelledby="chat-matches-title">
                            <h3 id="chat-matches-title">Matches</h3>
                            <div className="chat-matches">
                                {matches.length > 0 ? matches.map((match) => {
                                    const isActive = match.id === activeRecipientID;
                                    return (
                                        <button
                                            key={match.id}
                                            className={`chat-match${isActive ? " is-active" : ""}`}
                                            type="button"
                                            onClick={() => selectMatch(match)}
                                        >
                                            <img
                                                src={getAvatarURL(match.avatar)}
                                                alt=""
                                            />
                                            <span>
                                                <strong>{getMatchName(match)}</strong>
                                                <small>@{match.userName || `user${match.id}`}</small>
                                            </span>
                                        </button>
                                    );
                                }) : (
                                    <p className="chat-empty">No matches yet</p>
                                )}
                            </div>
                        </section>

                        <section className="chat-sidebar-section" aria-labelledby="chat-conversations-title">
                            <h3 id="chat-conversations-title">Conversations</h3>
                            <div className="chat-conversations">
                                {conversations.length > 0 ? conversations.map((conversation) => {
                                    const peerID = getPeerID(conversation, userID);
                                    const peerMatch = matchesByUserID.get(peerID);
                                    const isActive = conversation.id === activeConversationID;
                                    return (
                                        <button
                                            key={conversation.id}
                                            className={`chat-conversation${isActive ? " is-active" : ""}`}
                                            type="button"
                                            onClick={() => selectConversation(conversation)}
                                        >
                                            <span>{peerMatch ? getMatchName(peerMatch) : `User #${peerID}`}</span>
                                            <small>Open conversation</small>
                                        </button>
                                    );
                                }) : (
                                    <p className="chat-empty">No conversations yet</p>
                                )}
                            </div>
                        </section>
                    </>
                )}
            </aside>

            <section className="chat-panel" aria-label="Messages">
                <div className="chat-panel-header">
                    <div>
                        <p className="chat-kicker">Conversation</p>
                        <h2>{activeRecipientName}</h2>
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
