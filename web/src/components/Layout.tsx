import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="app">
      <nav className="navbar">
        <Link to="/library" className="nav-brand">ReadShelf</Link>
        <div className="nav-links">
          <Link to="/library">Library</Link>
          <Link to="/search">Search</Link>
          <Link to="/recall">AI Recall</Link>
          <button onClick={handleLogout} className="btn-link">Logout</button>
        </div>
      </nav>
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}
