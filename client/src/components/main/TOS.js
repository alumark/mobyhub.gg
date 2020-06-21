import React, { Component } from "react";

class TOS extends Component {
  render() {
    return (
      <div style={{ height: "75vh" }} className="container valign-wrapper">
        <div className="row">
          <div className="col s12 center-align">
            1. Do not mass share accounts, generally we are leanient with this rule, but if it gets excessive with the amount of IPs we will blacklist you.<br/>

            2. Don't sell accounts, you're allowed to sell keys, but not accounts. Selling an account can be problematic since they can be shared amongst multiple people.<br/>

            3. Don't compromise accounts. This is an immediate blacklist.<br/>
          </div>
        </div>
      </div>
    );
  }
}
export default TOS;