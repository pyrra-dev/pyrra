import React, {useEffect, useState} from 'react';
import './App.css';
import {Col, Container, Row} from 'react-bootstrap';

interface Objective {
    target: number
    window: string
}

interface Availability {
    percentage: number
    total: number
    errors: number
}

interface ErrorBudget {
    total: number
    max: number
    remaining: number
}

interface Valet {
    window: number
    volume: number
    errors: number
    availability: number
    budget: number
}

function App() {
    const [objective, setObjective] = useState<Objective | null>(null);
    const [availability, setAvailability] = useState<Availability | null>(null);
    const [errorBudget, setErrorBudget] = useState<ErrorBudget | null>(null);
    const [valets, setValets] = useState<Array<Valet>>([]);

    useEffect(() => {
        fetch('http://localhost:9099/objective.json')
            .then((resp) => resp.json())
            .then((data) => {
                setObjective(data.objective)
                setAvailability(data.availability)
                setErrorBudget(data.budget)
            })
    }, [])
    useEffect(() => {
        fetch('http://localhost:9099/objective/valet.json')
            .then((resp) => resp.json())
            .then((data) => setValets(data))
    }, [])

    return (
        <div className="App">
            <Container>
                <Row>
                    <Col>
                        <h1>My Important Service</h1>
                        <h4 className="text-muted">Important Service has to be 99% available in 30 days!</h4>
                    </Col>
                </Row>
                <br/><br/><br/>
                <Row>
                    <Col className="metric">
                        {availability != null ?
                            <div>
                                <h2 className={availability.percentage > 0 ? 'good' : 'bad'}>{(100 * availability.percentage).toFixed(3)}%</h2>
                                <h6 className="text-muted">Current</h6>
                            </div>
                            : <></>}
                    </Col>
                    <Col className="metric">
                        {objective != null ?
                            <div>
                                <h2>{100 * objective.target}%</h2>
                                <h6 className="text-muted">Goal in {objective.window}</h6>
                            </div>
                            : <></>}
                    </Col>
                    <Col className="metric">
                        {errorBudget != null ?
                            <div>
                                <h2 className={errorBudget.remaining > 0 ? 'good' : 'bad'}>{(100 * errorBudget.remaining).toFixed(3)}%</h2>
                                <h6 className="text-muted">Error Budget</h6>
                            </div>
                            : <></>}
                    </Col>
                </Row>
                <br/>
                <Row>
                    <img src="http://localhost:9099/objective/errorbudget.svg" alt=""/>
                </Row>
                <br/><br/>
                <Row>
                    <table className="table">
                        <thead>
                        <tr>
                            <th scope="col">Window</th>
                            <th scope="col">Volume</th>
                            <th scope="col">Errors</th>
                            <th scope="col">Availability</th>
                            <th scope="col">Error Budget</th>
                        </tr>
                        </thead>
                        <tbody>
                        {valets.map((v: Valet) => (
                            <tr key={v.window}>
                                <td>{v.window}</td>
                                <td>{Math.floor(v.volume)}</td>
                                <td>{Math.floor(v.errors)}</td>
                                <td>{(100 * v.availability).toFixed(3)}%</td>
                                <td>{(100 * v.budget).toFixed(3)}%</td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                </Row>
            </Container>
        </div>
    );
}

export default App;
