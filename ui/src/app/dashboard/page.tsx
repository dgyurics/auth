'use client'

import { useEffect, useState } from "react"
import { useRouter } from 'next/navigation'

export default function Dashboard() {
  const router = useRouter()
  const [sessions, setSessions] = useState([]); // State to store active sessions

  useEffect(() => {
    const websocketUrl = process.env.WS_URL
    if (!websocketUrl) {
      throw new Error('Websocket URL not found')
    }
    const newSocket = new WebSocket(websocketUrl)
    newSocket.onopen = () => {
      console.log('WebSocket connected', new Date().toUTCString())
    }
    newSocket.onmessage = (event) => {
      const receivedMessage = event.data
      // Update sessions state with new data
      setSessions(JSON.parse(receivedMessage) || [])
    }
    // FIXME Connection Close Frame	1703721217.7663352
    // handle connection close
    // why does logout-all close the websocket?
    newSocket.onerror = (event) => {
      console.error(event)
      router.push('/')
    }
    newSocket.onclose = () => {
      router.push('/')
      console.log('WebSocket disconnected', new Date().toUTCString())
    }
  }, [])

  const logout = async () => {
    try {
      await fetch(`${process.env.API_URL}/logout`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })
      router.push('/')
    } catch (err) {
      console.error(err)
    }
  }

  const logoutAll = async () => {
    try {
      await fetch(`${process.env.API_URL}/logout-all`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })
      router.push('/')
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <main className="flex flex-col justify-center items-center h-screen">
      <h1 className="text-4xl">Dashboard</h1>
      <div className="mt-4">
        <button
          type="button"
          onClick={logout}
          className="w-40 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
        >
          logout
        </button>
      </div>
      <div className="mt-4">
        <button
          type="button"
          onClick={logoutAll}
          className="w-40 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
        >
          logout-all
        </button>
      </div>
      <div className="mt-8 w-full max-w-4xl px-4 flex flex-col justify-center items-center">
        <h1 className="text-1xl">Active Sessions</h1>
        { sessions.map((session, index) => (<p>{session}</p>)) }
      </div>
    </main>
  )
}
