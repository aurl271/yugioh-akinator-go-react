import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import GameRoot from "./game/GameRoot";
import ConfirmPage from "./game/pages/confirm";
import FeedbackPage from "./game/pages/feedback";
import QuestionPage from "./game/pages/question";
import SetupPage from "./game/pages/setup";
import "./index.css";

function App() {

    return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/game/setup" replace />} />
        <Route path="/game" element={<GameRoot/>}>
            <Route index element={<Navigate to="setup" replace />} />
            <Route path="setup" element={<SetupPage />}/>
            <Route path="question" element={<QuestionPage />}/>
            <Route path="confirm" element={<ConfirmPage />}/>
            <Route path="feedback" element={<FeedbackPage />}/>
        </Route>
      </Routes>
    </BrowserRouter>
    )
}

export default App
