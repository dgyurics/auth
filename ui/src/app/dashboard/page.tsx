export default function Dashboard() {
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
