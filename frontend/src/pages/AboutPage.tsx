const AboutPage = () => {
  return (
    <div className="bg-gray-50 dark:bg-gray-900 dark:text-white min-h-screen py-12 transition-colors duration-200">
      <div className="container mx-auto px-4">
        <h1 className="text-3xl font-bold text-center mb-8">About Our Ticket Management System</h1>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-8">
          <h2 className="text-2xl font-semibold mb-4">Our Mission</h2>
          <p className="text-gray-700 dark:text-gray-300 mb-4">
            Our mission is to provide a streamlined, efficient ticket management solution that helps organizations
            track and resolve support issues quickly. We believe in creating tools that enhance productivity
            and improve communication between support teams and users.
          </p>
          <p className="text-gray-700 dark:text-gray-300">
            By centralizing ticket management, prioritizing issues, and facilitating clear communication,
            our system helps support teams deliver exceptional service while maintaining organization and accountability.
          </p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-8">
          <h2 className="text-2xl font-semibold mb-4">How It Works</h2>
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h3 className="text-xl font-medium mb-2">For Support Teams</h3>
              <ul className="list-disc pl-5 space-y-2 text-gray-700 dark:text-gray-300">
                <li>View and manage all tickets in a centralized dashboard</li>
                <li>Assign tickets to team members based on expertise</li>
                <li>Set priority levels to focus on critical issues</li>
                <li>Track ticket status from creation to resolution</li>
                <li>Communicate directly with users through ticket comments</li>
              </ul>
            </div>
            <div>
              <h3 className="text-xl font-medium mb-2">For Users</h3>
              <ul className="list-disc pl-5 space-y-2 text-gray-700 dark:text-gray-300">
                <li>Create tickets easily with a simple submission form</li>
                <li>Track the status of submitted tickets</li>
                <li>Receive updates when tickets are addressed</li>
                <li>Communicate with support staff through comments</li>
                <li>View history of past tickets and resolutions</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold mb-4">Our Team</h2>
          <p className="text-gray-700 dark:text-gray-300 mb-6">
            Our dedicated team of developers and support specialists work tirelessly to improve and maintain
            the Ticket Management System. With backgrounds in software development, customer support, and
            project management, our team brings diverse expertise to create a comprehensive solution.
          </p>

          <div className="grid md:grid-cols-3 gap-6 text-center">
            <div>
              <div className="w-24 h-24 bg-gray-300 dark:bg-gray-700 rounded-full mx-auto mb-3 flex items-center justify-center">
                <span className="text-gray-600 dark:text-gray-300 text-2xl">JP</span>
              </div>
              <h3 className="font-medium">John Peterson</h3>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Lead Developer</p>
            </div>
            <div>
              <div className="w-24 h-24 bg-gray-300 dark:bg-gray-700 rounded-full mx-auto mb-3 flex items-center justify-center">
                <span className="text-gray-600 dark:text-gray-300 text-2xl">SM</span>
              </div>
              <h3 className="font-medium">Sarah Miller</h3>
              <p className="text-gray-600 dark:text-gray-400 text-sm">UX Designer</p>
            </div>
            <div>
              <div className="w-24 h-24 bg-gray-300 dark:bg-gray-700 rounded-full mx-auto mb-3 flex items-center justify-center">
                <span className="text-gray-600 dark:text-gray-300 text-2xl">RJ</span>
              </div>
              <h3 className="font-medium">Robert Johnson</h3>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Support Manager</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AboutPage;