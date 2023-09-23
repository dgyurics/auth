'use client'

import { useState } from 'react'
import router from 'next/router'

export default function Home() {
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  })
  const [error, setError] = useState('')

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }))
    setError('')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await fetch(`${process.env.API_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      })

      if (response.ok) {
        router.push('/dashboard')
      } else {
        setError('Invalid username or password')
      }
    } catch (err) {
      setError('Network error')
    }
  }

  return (
    <main className="flex justify-center items-center h-screen">
      <form onSubmit={handleSubmit} className="w-1/4">
        <div className="mb-6">
          <input
            type="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            placeholder="username"
            className="w-full p-1 text-center border-b border-gray-400 focus:outline-none"
            required
          />
        </div>
        <div className="mb-8">
          <input
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            placeholder="password"
            className="w-full p-1 text-center border-b border-gray-400 focus:outline-none"
            required
          />
        </div>
        <div className="mb-4 flex justify-center">
          {error && (
            <div className="text-red-500 text-sm">{error}</div>
          )}
        </div>
        <div className="flex justify-center">
          <button
            type="submit"
            className="w-40 py-2 text-gray-400 border border-gray-400 focus:outline-none"
          >
            login
          </button>
        </div>
      </form>
    </main>
  )
}
