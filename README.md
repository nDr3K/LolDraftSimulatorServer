# League of Legends Draft Simulator (Backend)

This project is a Go-based server that powers the League of Legends Draft Simulator frontend. It handles lobby creation, manages draft states, and facilitates real-time communication using WebSockets. The server works seamlessly with the [React frontend](https://github.com/nDr3K/LolDraftSimulatorWebsite) to enable multiplayer and solo drafts.

## Features:
- **Lobby Creation**: Allows users to create and join lobbies for drafting.
- **Draft State Management**: The server keeps track of the draft's progress, including champion picks and bans.
- **Real-time Communication**: Utilizes WebSockets for real-time updates, ensuring smooth and dynamic gameplay interactions between players and spectators.
- **Fearless Mode Support**: Supports the fearless draft mode for a more intense experience.
- **Ban Style Options**: Manages different ban styles such as **SoloQ** and **Tournament-style** bans, based on the preferences set by the frontend.

## Live Demo:
For a fully integrated experience, try the live demo of the frontend at: [https://fearlessdraft.andreacannavo.com](https://fearlessdraft.andreacannavo.com).