'use client'

import { useEffect } from "react";

export default function Dashboard() {
  useEffect(() => {
    const websocketUrl = process.env.WS_URL
    if (!websocketUrl) {
      throw new Error('Websocket URL not found')
    }
    const newSocket = new WebSocket(websocketUrl)
    newSocket.onopen = () => {
      console.log('WebSocket connected')
    }
    newSocket.onmessage = (event) => {
      const receivedMessage = event.data
      console.log(receivedMessage)
    }        
  }, [])

  return (
    <main className="flex flex-col justify-center items-center h-screen">
      <h1 className="text-4xl">Dashboard</h1>
      <div className="mt-4">
        <button
          type="button"
          className="w-40 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
        >
          logout
        </button>
      </div>
      <div className="mt-4">
        <button
          type="button"
          className="w-40 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
        >
          logout all
        </button>
      </div>
    </main>
  );
}
