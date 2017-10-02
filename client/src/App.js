import React, { Component } from 'react';
import './App.css';

class App extends Component {

  state = {
    repos: {}
  }

  processMessage = (event) => {
    const data = JSON.parse(event.data);
    this.setState({
      repos: {
        ...this.state.repos,
        [data.slug]: data,
      },
    });
  }

  componentDidMount() {
    fetch("/repos").then((res) => res.json()).then((res) => {
      this.setState({repos: res})
    })
      .then(() => {

        this.ws = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws");
        this.ws.onmessage = e => this.processMessage(e);
      })
  }


  render() {

    const repos = Object.keys(this.state.repos).map((slug) => {
      const repo = this.state.repos[slug];
      return (<div key={repo.slug} className={`Repo Repo--${repo.last_build_state}`}>{repo.slug}</div>);
    })
    return (
      <div className="App">
        <div className="Repos">
          {repos}
        </div>
      </div>
    );
  }
}

export default App;
