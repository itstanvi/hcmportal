import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const isAdmin = user?.role === 'ADMIN';

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="layout">
      <aside className="sidebar">
        <h1>HCM Console</h1>
        <nav>
          <NavLink to="/employees" className={({ isActive }) => (isActive ? 'active' : '')}>
            Employees
          </NavLink>
          <NavLink to="/employees/new" className={({ isActive }) => (isActive ? 'active' : '')}>
            Add Employee
          </NavLink>
          {isAdmin && (
            <NavLink to="/fields" className={({ isActive }) => (isActive ? 'active' : '')}>
              Field Configuration
            </NavLink>
          )}
        </nav>
        <div className="user">
          <div className="name">{user?.name}</div>
          <div className="role">{user?.email} · {user?.role}</div>
          <button onClick={handleLogout}>Sign out</button>
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
