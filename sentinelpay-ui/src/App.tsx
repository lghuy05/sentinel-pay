import { NavLink, Outlet } from "react-router-dom";

const App = () => {
  return (
    <div className="app-shell">
      <header className="navbar">
        <div className="navbar-inner">
          <div className="brand">
            <span className="brand-mark"></span>
            SentinelPay Ops Console
          </div>
          <nav className="nav-links">
            <NavLink to="/" className={({ isActive }) => (isActive ? "active" : undefined)}>
              Dashboard
            </NavLink>
            <NavLink to="/accounts" className={({ isActive }) => (isActive ? "active" : undefined)}>
              Accounts
            </NavLink>
            <NavLink to="/transactions" className={({ isActive }) => (isActive ? "active" : undefined)}>
              Simulate Transaction
            </NavLink>
            <NavLink to="/decisions" className={({ isActive }) => (isActive ? "active" : undefined)}>
              Fraud Decisions
            </NavLink>
            <NavLink to="/feedback" className={({ isActive }) => (isActive ? "active" : undefined)}>
              Feedback
            </NavLink>
            <NavLink to="/status" className={({ isActive }) => (isActive ? "active" : undefined)}>
              System Status
            </NavLink>
          </nav>
        </div>
      </header>

      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
};

export default App;
