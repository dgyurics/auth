'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'

export default function Home() {
  const router = useRouter()
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

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    await handleLogin()
  }

  const handleLogin = async () => {
    try {
      const response = await fetch(`${process.env.API_URL}/login`, {
        method: 'POST',
        credentials: 'include',
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

  const handleRegister = async () => {
    try {
      const response = await fetch(`${process.env.API_URL}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      })

      if (response.ok) {
        router.push('/dashboard')
      } else if (response.status === 409) {
        setError('Username already exists')
      } else {
        setError('Invalid username or password')
      }
    } catch (err) {
      setError('Network error')
    }
  }

  return (
    <main className="flex justify-center items-center h-screen">
      <form className="w-1/4" onSubmit={handleSubmit}>
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
        <div className="flex justify-center space-x-4"> {/* Added space-x-4 for spacing */}
          <button
            type="submit"
            className="w-32 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
          >
            Login
          </button>
          <button
            type="button"
            onClick={() => handleRegister()}
            className="w-32 py-2 text-gray-600 border border-gray-600 hover:text-gray-400 hover:border-gray-400 focus:outline-none transition duration-300"
          >
            Register
          </button>
        </div>
      </form>
    </main>
  )
}
